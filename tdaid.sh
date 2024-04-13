#!/bin/bash

usage="usage: tdaid.sh { -b branch} { -I container_image } {-c | -t | -s sysmsg } [outputfile1] [outputfile2] ...
    -b:  branch name
    -c:  write code
    -I:  container image name
    -t:  write tests
    -s:  execute custom sysmsg
"

cmdline="$0 $@"

# parse command line options
while getopts "b:cI:s:tZ" opt
do
    case $opt in
        b)  branch=$OPTARG
            ;;
        c)  mode=code
            ;;
        I)  container_image=$OPTARG
            ;;
        s)  mode=custom
            sysmsgcustom=$OPTARG
            ;;
        t)  mode=tests
            ;;
        Z)  inContainer=1
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
    go test -v -timeout 1m
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

outfns="$@"

infns=$(find * -type f)

infnsComma=$(echo $infns | tr ' ' ',')
outfnsComma=$(echo $outfns | tr ' ' ',')

sysmsgcode="You are an expert Go programmer.  Write or fix code in
[$outfns] to make the tests pass.  Do not ask me to do things -- take
care of it yourself.  I am giving you all relevant files.  Do not mock
the results.  Write complete, production-quality code.  Do not write
stubs.  Do not omit code -- provide the complete file each time.  Do
not enclose backticks in string literals -- you can't escape backticks
in Go, so you'll need to build string literals with embedded backticks
by using string concatenation. Include comments and follow the Go
documentation conventions.  If you see an error in the tests, say
TESTERROR on a line by itself and suggest a fix."

sysmsgtest="You are an expert Go programmer.  Appends tests to
[$outfns] to make the code more robust.  Do not alter or insert before
existing tests.  Do not inline multiline test data in Go files -- put
test data in the given output data files.  Do not enclose backticks in
string literals -- you can't escape backticks in Go, so you'll need to
build string literals with embedded backticks by using string
concatenation. If you see an error in the code, say CODEERROR on a
line by itself and suggest a fix."

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

# make a stamp file dated at time zero
touch -t 197001010000 /tmp/$$.stamp

# To reduce build time, we run tidy in the container and commit the
# container with a temporary name, then use that temporary container
# in the test loop, then delete it after the run.
tmp_container_image=$container_image-tmp-delete-me
docker run \
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
        -v $0:/tmp/tdaid \
        -w /mnt \
        $tmp_container_image /tmp/tdaid -Z 2>&1 | tee /tmp/$$.test

    case $mode in
        code)   sysmsg=$sysmsgcode
                # if tests pass, exit
                if ! grep -q "FAIL" /tmp/$$.test
                then
                    grok chat /tmp/$$.chat -i $infnsComma -s "Recommend additional tests to improve coverage and robustness of code." < /tmp/$$.test
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
        grok chat /tmp/$$.chat -i $infnsComma -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    else
        grok chat /tmp/$$.chat -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    fi
    set +x

    # look for TESTERROR or CODEERROR
    errcount=$(egrep "^(TESTERROR|CODEERROR)$" /tmp/$$.chat | wc -l)
    if [ $errcount -gt 3 ]
    then
        break
    fi

    # test for vet errors -- if found, don't commit
    go vet || continue

    # commit new code or tests
    set -x
    git add $infns $outfns 
    grok commit > /tmp/$$.commit
    git commit -F /tmp/$$.commit
    set +x

    sleep 1
done

# cleanup
docker rmi $tmp_container_image

echo "# to squash and merge the dev branch into main or master, run the following commands:"
echo "git checkout main || git checkout master"
echo "git merge --squash $branch"
echo "git commit"
