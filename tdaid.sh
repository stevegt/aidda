#!/bin/bash

usage="usage: tdaid.sh [-c | -t] [outputfile1] [outputfile2] ...
    -c:  write code
    -t:  write tests
"

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

sysmsgcode="You are an expert Go programmer.  Write code to make the
tests pass.  Do not mock the results.  Write production-quality code.
If you see an error in the tests, say TESTERROR and suggest a fix."

sysmsgtest="You are an expert Go programmer.  Write tests to make the
code more robust.  If you see an error in the code, say CODEERROR and
suggest a fix."

case $mode in
    code)   sysmsg=$sysmsgcode
            ;;
    tests)  sysmsg=$sysmsgtest
            ;;
esac

# loop until tests pass
while true
do
    # run tests
    (
        go mod tidy
        go test -v 
    ) 2>&1 | tee /tmp/$$.test

    # look for TESTERROR or CODEERROR
    if grep -q "TESTERROR" /tmp/$$.test
    then
        break
    fi
    if grep -q "CODEERROR" /tmp/$$.test
    then
        break
    fi

    # if tests pass, exit
    if ! grep -q "FAIL" /tmp/$$.test
    then
        set -x
        grok chat /tmp/$$.chat -i $infnsComma -s "Suggest further tests to make the code more robust." < /tmp/$$.test
        set +x
        break
    fi

    # get new code from grokker
    set -x
    grok chat /tmp/$$.chat -i $infnsComma -o $outfnsComma -s "$sysmsg" < /tmp/$$.test
    set +x
    sleep 1
done
