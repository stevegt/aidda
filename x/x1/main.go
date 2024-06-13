package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const usage = `usage: aidda.go { -b branch} { -I container_image } {-a sysmsg | -c | -t | -s sysmsg } [-A 'go test' args ] [ -p input_patterns_file ] [outputfile1] [outputfile2] ...
	modes:
	-a:  skip tests and provide advice
	-c:  write code
	-t:  write tests
	-s:  execute custom sysmsg

	-A:  extra arguments to pass to 'go test'
	-b:  branch name
	-C:  continue chat from existing chatfile
	-I:  container image name
	-p:  file containing input filename patterns
	-T:  test timeout e.g. '1m'
`

func main() {
	var testArgs, branch, chatfile, mode, containerImage, sysmsgcustom, inpatfn, inContainer string
	var outfns []string

	flag.StringVar(&testArgs, "A", "./...", "extra arguments to pass to 'go test'")
	flag.StringVar(&branch, "b", "", "branch name")
	flag.StringVar(&chatfile, "C", "/tmp/aidda-chat", "continue chat from existing chatfile")
	flag.StringVar(&containerImage, "I", "", "container image name")
	flag.StringVar(&sysmsgcustom, "s", "", "custom sysmsg")
	flag.StringVar(&inpatfn, "p", "", "file containing input filename patterns")
	flag.StringVar(&inContainer, "Z", "", "inContainer option")

	flag.Parse()

	outfns = flag.Args()

	fmt.Printf("aidda.go %v\n", os.Args)
	// cmdline := strings.Join(os.Args, " ")

	if inContainer != "" {
		runInContainer(inContainer)
		return
	}

	stampFile := "/tmp/stamp"
	createStampFile(stampFile, chatfile)

	infns := getInputFiles(inpatfn, stampFile)

	if mode == "advice" {
		if sysmsgcustom == "" {
			fmt.Println("error: sysmsg required")
			fmt.Println(usage)
			os.Exit(1)
		}
		runAdviceMode(chatfile, infns, sysmsgcustom)
		return
	}

	if mode == "" || branch == "" || containerImage == "" || len(outfns) < 1 {
		fmt.Printf("mode: %s\nbranch: %s\ncontainer_image: %s\nargs: %d\n", mode, branch, containerImage, len(outfns))
		fmt.Println(usage)
		os.Exit(1)
	}

	switch mode {
	case "code":
		sysmsg := "You are an expert Go programmer. Write, add, or fix the target code in " + strings.Join(outfns, ",") + " to make the tests pass. ..."
		runCodeMode(branch, containerImage, sysmsg, chatfile, infns, outfns)
	case "tests":
		sysmsg := "You are an expert Go programmer. Append tests to " + strings.Join(outfns, ",") + " to make the code more robust. ..."
		runTestsMode(branch, containerImage, sysmsg, chatfile, infns, outfns)
	case "custom":
		sysmsg := sysmsgcustom
		runCustomMode(branch, containerImage, sysmsg, chatfile, infns, outfns)
	default:
		fmt.Println(usage)
		os.Exit(1)
	}
}

func createStampFile(stampFile, chatfile string) {
	runCommand(fmt.Sprintf("touch -t 197001010000 %s", stampFile))
	if _, err := os.Stat(chatfile); err == nil {
		runCommand(fmt.Sprintf("touch -r %s %s", chatfile, stampFile))
	}
}

func getInputFiles(inpatfn, stampFile string) string {
	if inpatfn != "" {
		runCommand("set -ex")
		infns := ""
		// Handle reading patterns and finding files
		runCommand("set +x")
		return infns
	}
	infns := runCommand(fmt.Sprintf("find * -type f -newer %s", stampFile))
	return infns
}

func runAdviceMode(chatfile, infns, sysmsgcustom string) {
	runCommand(fmt.Sprintf("grok chat %s -i %s -s \"%s\" < /dev/null", chatfile, infns, sysmsgcustom))
}

func runCodeMode(branch, containerImage, sysmsg, chatfile, infns string, outfns []string) {
	if !isRepoClean() {
		fmt.Println("error: changes must be committed")
		os.Exit(1)
	}

	curbranch := getCurrentBranch()
	checkoutBranch(branch)
	mergeBranch(curbranch)

	tmpContainerImage := containerImage + "-tmp-delete-me"
	cleanupContainers(tmpContainerImage)
	tidyAndCommitContainer(containerImage, tmpContainerImage)

	startTime := time.Now()
	for {
		if time.Since(startTime) > 20*time.Minute {
			fmt.Println("error: time limit exceeded")
			break
		}

		runTests(tmpContainerImage, testArgs)

		if mode == "code" && testsPass() {
			recommendAdditionalTests(chatfile, infns, outfns)
			break
		}

		newFiles := getUpdatedFiles(infns, outfns, stampFile)
		updateFilesFromGrok(chatfile, sysmsg, newFiles, outfns)

		if errorOccurred(chatfile) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if goVet() {
		commitChanges(infns, outfns)
		printSquashAndMergeInstructions(branch)
	}

	cleanupContainers(tmpContainerImage)
}

func runTestsMode(branch, containerImage, sysmsg, chatfile, infns string, outfns []string) {
	// Similar implementation to runCodeMode
}

func runCustomMode(branch, containerImage, sysmsg, chatfile, infns string, outfns []string) {
	// Similar implementation to runCodeMode
}

func runCommand(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	return string(out)
}

func isRepoClean() bool {
	status := runCommand("git status --porcelain")
	return status == ""
}

func getCurrentBranch() string {
	return runCommand("git branch --show-current")
}

func checkoutBranch(branch string) {
	runCommand(fmt.Sprintf("git checkout %s", branch))
}

func mergeBranch(branch string) {
	runCommand(fmt.Sprintf("git merge --commit %s", branch))
}

func recommendAdditionalTests(chatfile, infns string, outfns []string) {
	runCommand(fmt.Sprintf("grok chat %s -i %s -s \"Recommend additional tests to improve coverage and robustness of code.\" < /tmp/test", chatfile, infns))
}

func getUpdatedFiles(infns string, outfns []string, stampFile string) string {
	newFiles := ""
	// logic to get updated files
	return newFiles
}

func updateFilesFromGrok(chatfile, sysmsg, newFiles string, outfns []string) {
	runCommand(fmt.Sprintf("grok chat %s -i %s -o %s -s \"%s\" < /tmp/test", chatfile, newFiles, strings.Join(outfns, ","), sysmsg))
}

func errorOccurred(chatfile string) bool {
	out := runCommand(fmt.Sprintf("egrep '^\\s*(TESTERROR|CODEERROR)\\s*$' %s | wc -l", chatfile))
	return strings.TrimSpace(out) != "0"
}

func goVet() bool {
	err := exec.Command("go", "vet").Run()
	return err == nil
}

func commitChanges(infns string, outfns []string) {
	runCommand(fmt.Sprintf("git add %s %s", infns, strings.Join(outfns, " ")))
	commitMsg := runCommand("grok commit")
	tmpCommitFile := "/tmp/commit"
	runCommand(fmt.Sprintf("echo \"%s\" > %s", commitMsg, tmpCommitFile))
	runCommand(fmt.Sprintf("git commit -F %s", tmpCommitFile))
}

func printSquashAndMergeInstructions(branch string) {
	fmt.Println("# to squash and merge the dev branch into main or master, run the following commands:")
	fmt.Println("git checkout main || git checkout master")
	fmt.Println(fmt.Sprintf("git merge --squash %s", branch))
	fmt.Println("git commit")
}
