#!/bin/bash

usage="usage: aidda.sh { -b branch} { -I container_image } {-a sysmsg | -c | -t | -s sysmsg } [-r] [-A 'go test' args ] [ -i input_files ] [outputfile1] [outputfile2] ...
    modes:
    -a:  skip tests and provide advice
    -c:  write code
    -t:  write tests
    -s:  execute custom sysmsg

    -A:  extra arguments to pass to 'go test'
    -b:  branch name
    -C:  continue chat from existing chatfile
    -I:  container image name
    -i:  input files, comma separated as in 'grok chat -i'
    -r:  run tests with -race
    -T:  test timeout e.g. '1m'
"
echo "aidda.sh $@"
cmdline="$0 $@"

# parse command line options
containerArgs=" "
chatfile=/tmp/aidda-$$.chat
while getopts "A:a:b:C:cI:i:s:tZ:" opt
do
    case $opt in
        A)  containerArgs="$containerArgs $OPTARG"
            ;;
        a)  mode=advice
            sysmsgcustom=$OPTARG
            ;;
        b)  branch=$OPTARG
            ;;
        C)  chatfile=$OPTARG
            ;;
        c)  mode=code
            ;;
        I)  container_image=$OPTARG
            ;;
        i)  infnsComma=$OPTARG
            ;;
        s)  mode=custom
            sysmsgcustom=$OPTARG
            ;;
        t)  mode=tests
            ;;
        Z)  inContainer="$OPTARG"
            ;;
        *)  echo "$usage"
            exit 1
            ;;
    esac
done
shift $((OPTIND - 1))

if [ -n "$inContainer" ]
then
    set -ex
    go mod tidy
    golint 
    go test -v $inContainer ./...
    exit 0
fi

# make a stamp file dated at time zero
touch -t 197001010000 /tmp/$$.stamp
# if chat file exists, set stamp time to chat file time
if [ -e $chatfile ]
then
    touch -r $chatfile /tmp/$$.stamp
fi

outfns="$@"
outfnsComma=$(echo $outfns | tr ' ' ',')

if [ -z "$infnsComma" ]
then
    infns=$(find * -type f -newer /tmp/$$.stamp)
    infnsComma=$(echo $infns | tr ' ' ',')
fi

if [ "$mode" == "advice" ] 
then
    if [ -z "$sysmsgcustom" ] 
    then
        echo "error: sysmsg required"
        echo "$usage"
        exit 1
    fi
    set -x
    cmd="grok chat $chatfile"
    if [ -n "$infnsComma" ]
    then
        cmd="$cmd -i $infnsComma"
    fi
    msgflag="-s"
    if [ -e $chatfile ]
    then
        msgflag="-m"
    fi
    $cmd $msgflag "$sysmsgcustom" < /dev/null
    exit 0
fi

if [ -z "$mode" ] || [ -z "$branch" ] || [ -z "$container_image" ] || [ $# -lt 1 ]
then
    echo mode: $mode
    echo branch: $branch
    echo container_image: $container_image
    echo $#: $#
    echo "$@"
    echo "$usage"
    exit 1
fi

sysmsgcode="You are an expert Go programmer.  Write, add, or fix the
target code in [$outfns] to make the tests pass.  In case of conflict
between tests and target code, consider the tests to be correct.
Create any missing types, methods, or fields referenced by the tests.
I am giving you all relevant files. Do not mock the results.  Write
complete, production-quality code.  Do not write stubs.  Do not omit
code -- provide the complete file each time.  Do not enclose backticks
in string literals -- you can't escape backticks in Go, so you'll need
to build string literals with embedded backticks by using string
concatenation. Include comments and follow the Go documentation
conventions.  If you are unable to follow these instructions, say
TESTERROR on a line by itself and suggest a fix."

sysmsgtest="You are an expert Go programmer.  Appends tests to
[$outfns] to make the code more robust.  Do not alter or insert before
existing tests.  Do not inline multiline test data in Go files -- put
test data in the given output data files.  Do not enclose backticks in
string literals -- you can't escape backticks in Go, so you'll need to
build string literals with embedded backticks by using string
concatenation. If you see an error in the code or need me to do
anything, say CODEERROR on a line by itself and suggest a fix."

# ensure repo is clean
stat=$(git status --porcelain)
if [ -n "$stat" ]
then
    echo "error: changes must be committed"
    exit 1
fi

# get current branch name
curbranch=$(git branch --show-current)

# checkout dev branch
set -ex
git checkout $branch
set +ex

# merge from curbranch
set -ex
git merge --commit $curbranch
set +ex

tmp_container_image=$container_image-tmp-delete-me
# cleanup any previous containers
docker container ls -a -f label=aidda-delete-me -q | xargs docker stop
sleep 1
docker rmi $tmp_container_image 
docker container ls -a -f label=aidda-delete-me -q | xargs docker rm
docker image ls -a -f label=aidda-delete-me -q | xargs docker rmi

# To reduce build time, we run tidy in the container and commit the
# container with a temporary name, then use that temporary container
# in the test loop, then delete it after the run.
docker run \
    --label aidda-delete-me \
    -v $(pwd):/mnt \
    -w /mnt \
    $container_image go mod tidy
docker commit $(docker ps -lq) $tmp_container_image

# loop until tests pass or timeout
startTime=$(date +%s)
while true
do
    # limit runtime to 20 minutes
    endTime=$(date +%s)
    if [ $(($endTime - $startTime)) -gt 1200 ]
    then
        echo "error: time limit exceeded"
        break
    fi

    # run tests
    docker run --rm \
        -v $(pwd):/mnt \
        -v $0:/tmp/aidda \
        -w /mnt \
        $tmp_container_image /tmp/aidda -Z "$containerArgs" 2>&1 | tee /tmp/$$.test

    case $mode in
        code)   sysmsg=$sysmsgcode
                # if tests pass, exit
                if ! grep -q "FAIL" /tmp/$$.test
                then
                    grok chat $chatfile -i $infnsComma -s "Recommend additional tests to improve coverage and robustness of code." < /tmp/$$.test
                    break
                fi
                ;;
        tests)  sysmsg=$sysmsgtest
                # if tests fail, exit
                if grep -q "FAIL" /tmp/$$.test
                then
                    break
                fi
                ;;
        custom) sysmsg=$sysmsgcustom
                # if tests pass, exit
                if ! grep -q "FAIL" /tmp/$$.test
                then
                    break
                fi
                ;;
    esac

    # only include input files that have been updated since the last run
    newfns=""
    for infn in $infns
    do
        # skip output files
        for outfn in $outfns
        do
            if [ "$infn" = "$outfn" ]
            then
                continue 2
            fi
        done
        if [ "$infn" -nt /tmp/$$.stamp ]
        then
            newfns="$newfns $infn"
        fi
    done
    newfnsComma=$(echo $newfns | tr ' ' ',')
    touch /tmp/$$.stamp

    # get new code or tests from grokker
    set -x
    if [ "$newfnsComma" != "" ]
    then
        grok chat $chatfile -i $infnsComma -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    else
        grok chat $chatfile -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    fi
    set +x

    # look for TESTERROR or CODEERROR
    errcount=$(egrep "^\s*(TESTERROR|CODEERROR)\s*$" $chatfile | wc -l)
    # try to fix the error N times before giving up
    if [ $errcount -gt 1 ]
    then
        break
    fi

    sleep 1
done

# test for vet errors -- if found, don't commit
if go vet 
then
    # commit new code or tests
    set -x
    git add $infns $outfns 
    grok commit > /tmp/$$.commit
    git commit -F /tmp/$$.commit
    set +x

    echo "# to squash and merge the dev branch into main or master, run the following commands:"
    echo "git checkout main || git checkout master"
    echo "git merge --squash $branch"
    echo "git commit"
fi

# cleanup
docker rmi $tmp_container_image
