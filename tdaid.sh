#!/bin/bash

usage="usage: tdaid.sh [-c | -t] [outputfile1] [outputfile2] ...
    -c:  write code
    -t:  write tests
"

cmdline="$0 $@"

# parse command line options
unset mode
while getopts "ct" opt
do
    case $opt in
        c)  mode=code
            shift
            ;;
        t)  mode=tests
            shift
            ;;
    esac
done
if [ -z "$mode" ]
then
    echo "$usage"
    exit 1
fi

outfns="$@"

infns=$(find * -type f)

infnsComma=$(echo $infns | tr ' ' ',')
outfnsComma=$(echo $outfns | tr ' ' ',')

sysmsgcode="You are an expert Go programmer.  Write or fix code in
[$outfns] to make the tests pass.  Do not mock the results.  Write
production-quality code. Do not attempt to enclose backticks in string
literals -- you can't escape backticks in Go, so you'll need to build
string literals with embedded backticks by using string concatenation.
If you see an error in the tests, say TESTERROR on a line by itself
and suggest a fix."

sysmsgtest="You are an expert Go programmer.  Add tests to [$outfns]
to make the code more robust.  Do not alter existing tests.  Put test
data in non-Go files if given.  Do not attempt to enclose backticks in
string literals -- you can't escape backticks in Go, so you'll need to
build string literals with embedded backticks by using string
concatenation.  If you see an error in the code, say CODEERROR on a
line by itself and suggest a fix."

# ensure all input and output files are committed
stat=$(git status --porcelain $infns $outfns)
if [ -n "$stat" ]
then
    echo "error: input and output files must be committed"
    exit 1
fi

# start a temporary git branch
set -ex
git checkout -b tdaid_$$ 
set +ex

# loop until tests pass
first=1
while true
do
    # run tests
    (
        go mod tidy
        go test -v 
    ) 2>&1 | tee /tmp/$$.test

    case $mode in
        code)   sysmsg=$sysmsgcode
                # if tests pass, exit
                if ! grep -q "FAIL" /tmp/$$.test
                then
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
    esac

    # get new code or tests from grokker
    if [ $first -eq 1 ]
    then
        # include input files in the first run
        grok chat /tmp/$$.chat -i $infnsComma -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
        first=0
    else
        # do not include input files in subsequent runs, since they are
        # already included in the chat file
        grok chat /tmp/$$.chat -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    fi

    # commit new code or tests
    set -ex
    git add $infns $outfns 
    git commit -m "$cmdline" 
    set +ex

    # look for TESTERROR or CODEERROR
    if egrep "^(TESTERROR|CODEERROR)$" /tmp/$$.chat
    then
        break
    fi

    sleep 1
done

echo "# to squash and merge the temporary branch into master or main and delete the temporary branch, run the following commands:"
echo "git checkout main || git checkout master"
echo "git merge --squash tdaid_$$"
echo "git commit"
echo "git branch -d tdaid_$$"
