package x3

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/emersion/go-message/mail"
	"github.com/fsnotify/fsnotify"
	"github.com/google/shlex"
	"github.com/stevegt/envi"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/grokker/v3/core"
	"github.com/stevegt/grokker/v3/util"
)

// RunTee runs a command in the shell, with stdout and stderr tee'd to the terminal
func RunTee(command string) (stdout, stderr []byte, rc int, err error) {
	defer Return(&err)
	// shlex the command to get the command and args
	parts, err := shlex.Split(command)
	Ck(err)
	var args []string
	var cmd string
	if len(parts) > 1 {
		cmd = parts[0]
		args = parts[1:]
	} else {
		cmd = parts[0]
	}
	// create the command
	cobj := exec.Command(cmd, args...)
	// connect stdin to the terminal
	cobj.Stdin = os.Stdin

	// create a tee for stdout
	stdoutPipe, err := cobj.StdoutPipe()
	Ck(err)
	stdoutTee := io.TeeReader(stdoutPipe, os.Stdout)
	// create a tee for stderr
	stderrPipe, err := cobj.StderrPipe()
	Ck(err)
	stderrTee := io.TeeReader(stderrPipe, os.Stderr)
	// read the stdout in a goroutine
	go func() {
		stdout, err = ioutil.ReadAll(stdoutTee)
		Ck(err)
	}()

	// read the stderr in a goroutine
	go func() {
		stderr, err = ioutil.ReadAll(stderrTee)
		Ck(err)
	}()
	// wait for goroutines to get started
	// XXX use a waitgroup instead
	time.Sleep(100 * time.Millisecond)

	// start the command
	err = cobj.Start()
	Ck(err)
	// wait for the command to finish
	err = cobj.Wait()
	Ck(err)
	// get the return code
	rc = cobj.ProcessState.ExitCode()
	return
}

// Run runs a command in the shell, returning stdout, stderr, and rc
func Run(command string, stdin []byte) (stdout, stderr []byte, rc int, err error) {
	defer Return(&err)
	// shlex the command to get the command and args
	parts, err := shlex.Split(command)
	Ck(err)
	var args []string
	var cmd string
	if len(parts) > 1 {
		cmd = parts[0]
		args = parts[1:]
	} else {
		cmd = parts[0]
	}
	// create the command
	cobj := exec.Command(cmd, args...)
	// create a pipe for stdin
	stdinPipe, err := cobj.StdinPipe()
	Ck(err)
	// create a pipe for stdout
	stdoutPipe, err := cobj.StdoutPipe()
	Ck(err)
	// create a pipe for stderr
	stderrPipe, err := cobj.StderrPipe()
	Ck(err)
	// start the command
	err = cobj.Start()
	Ck(err)
	// pipe stdin to the command in a goroutine
	go func() {
		_, err = stdinPipe.Write(stdin)
		Ck(err)
		stdinPipe.Close()
	}()
	// read the stdout in a goroutine
	go func() {
		stdout, err = ioutil.ReadAll(stdoutPipe)
		Ck(err)
	}()
	// read the stderr in a goroutine
	go func() {
		stderr, err = ioutil.ReadAll(stderrPipe)
		Ck(err)
	}()
	// wait for the command to finish
	err = cobj.Wait()
	Ck(err)
	// get the return code
	rc = cobj.ProcessState.ExitCode()
	return
}

// RunInteractive runs a command in the shell, with stdio connected to the terminal
func RunInteractive(command string) (rc int, err error) {
	defer Return(&err)
	// shlex the command to get the command and args
	parts, err := shlex.Split(command)
	Ck(err)
	var args []string
	var cmd string
	if len(parts) > 1 {
		cmd = parts[0]
		args = parts[1:]
	} else {
		cmd = parts[0]
	}
	// create the command
	cobj := exec.Command(cmd, args...)
	// connect the stdio to the terminal
	cobj.Stdin = os.Stdin
	cobj.Stdout = os.Stdout
	cobj.Stderr = os.Stderr
	// start the command
	err = cobj.Start()
	Ck(err)
	// wait for the command to finish
	err = cobj.Wait()
	Ck(err)
	// get the return code
	rc = cobj.ProcessState.ExitCode()
	return
}

/*
- while true
	- git commit
	- present user with an editor buffer where they can type a natural language instruction
	- send that along with all files to GPT API
		- filter out files using .aidda/ignore
	- save returned files over top of the existing files
	- run 'git difftool' with vscode as in https://www.roboleary.net/vscode/2020/09/15/vscode-git.html
	- open diff tool in editor so user can selectively choose and edit changes
	- run go test -v
	- include test results in the prompt file
*/

func Start(args ...string) {
	base := args[0]
	err := os.Chdir(filepath.Dir(base))
	Ck(err)

	// ensure there is a .git directory
	_, err = os.Stat(".git")
	Ck(err)

	// ensure there is a .grok file
	_, err = os.Stat(".grok")
	Ck(err)

	// generate a filename for the prompt file
	dir := Spf("%s/.aidda", filepath.Dir(base))
	err = os.MkdirAll(dir, 0755)
	Ck(err)
	fn := Spf("%s/prompt", dir)

	// open or create a grokker db
	g, lock, err := core.LoadOrInit(base, "gpt-4o")
	Ck(err)
	defer lock.Unlock()

	// commit the current state
	err = commit(g)
	Ck(err)

	// loop forever
	done := false
	for !done {
		done, err = loop(g, fn)
		Ck(err)
		time.Sleep(3 * time.Second)
	}
}

// Prompt is a struct that represents a prompt
type Prompt struct {
	In  []string
	Out []string
	Txt string
}

// NewPrompt opens or creates a prompt object
func NewPrompt(path string) (p *Prompt, err error) {
	defer Return(&err)
	// check if the file exists
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err = createPromptFile(path)
		Ck(err)
	} else {
		Ck(err)
	}
	p, err = readPrompt(path)
	Ck(err)
	return
}

// readPrompt reads a prompt file
func readPrompt(path string) (p *Prompt, err error) {
	p = &Prompt{}
	// parse the file as a mail message
	file, err := os.Open(path)
	Ck(err)
	defer file.Close()
	mr, err := mail.CreateReader(file)
	Ck(err)
	// read the message header
	header := mr.Header
	inStr := header.Get("In")
	outStr := header.Get("Out")
	p.In = strings.Split(inStr, ", ")
	p.Out = strings.Split(outStr, ", ")
	// read the message body
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		Ck(err)
		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			// prompt text is in the body
			buf, err := io.ReadAll(part.Body)
			Ck(err)
			// trim leading and trailing whitespace
			txt := strings.TrimSpace(string(buf))
			p.Txt = string(txt)
		case *mail.AttachmentHeader:
			// XXX keep this here because we might perhaps use
			// attachments in the future for e.g. test results
			filename, err := h.Filename()
			Ck(err)
			fmt.Printf("Got attachment: %v\n", filename)
		}
	}
	return
}

// createPromptFile creates a new prompt file
func createPromptFile(path string) (err error) {
	defer Return(&err)
	file, err := os.Create(path)
	Ck(err)
	defer file.Close()

	// get the list of files to process
	inFns, err := getFiles()
	outFns := inFns[:]
	inStr := strings.Join(inFns, ", ")
	outStr := strings.Join(outFns, ", ")

	// create headers
	hmap := map[string][]string{
		"In":  []string{inStr},
		"Out": []string{outStr},
	}
	h := mail.HeaderFromMap(hmap)

	// create mail writer
	mw, err := mail.CreateSingleInlineWriter(file, h)
	Ck(err)
	// Write the body
	io.WriteString(mw, "# enter prompt here")

	return
}

// ask asks the user a question and gets a response
func ask(question, deflt string, others ...string) (response string, err error) {
	defer Return(&err)
	var candidates []string
	candidates = append(candidates, strings.ToUpper(deflt))
	for _, o := range others {
		candidates = append(candidates, strings.ToLower(o))
	}
	for {
		fmt.Printf("%s [%s]: ", question, strings.Join(candidates, "/"))
		reader := bufio.NewReader(os.Stdin)
		response, err = reader.ReadString('\n')
		Ck(err)
		response = strings.TrimSpace(response)
		if response == "" {
			response = deflt
		}
		if len(others) == 0 {
			// if others is empty, return the response without
			// checking candidates
			return
		}
		// check if the response is in the list of candidates
		for _, c := range candidates {
			if strings.ToLower(response) == strings.ToLower(c) {
				return
			}
		}
	}
}

func loop(g *core.Grokker, promptfn string) (done bool, err error) {
	defer Return(&err)

	p, err := getPrompt(promptfn)
	Ck(err)
	spew.Dump(p)

	err = getChanges(g, p)
	Ck(err)

	err = runDiff()
	Ck(err)

	err = runTest(promptfn)
	Ck(err)

	err = commit(g)
	Ck(err)

	return
}

func runTest(promptfn string) (err error) {
	defer Return(&err)
	Pf("Running tests\n")

	// run go test -v
	stdout, stderr, _, _ := RunTee("go test -v")

	// append test results to the prompt file
	fh, err := os.OpenFile(promptfn, os.O_APPEND|os.O_WRONLY, 0644)
	Ck(err)
	_, err = fh.WriteString(Spf("\n\nstdout:\n%s\n\nstderr:%s\n\n", stdout, stderr))
	Ck(err)
	fh.Close()
	return err
}

func runDiff() (err error) {
	defer Return(&err)
	// run difftool
	difftool := envi.String("AIDDA_DIFFTOOL", "git difftool")
	Pf("Running difftool %s\n", difftool)
	var rc int
	rc, err = RunInteractive(difftool)
	Ck(err)
	Assert(rc == 0, "difftool failed")
	return err
}

func getChanges(g *core.Grokker, p *Prompt) (err error) {
	defer Return(&err)
	Pf("getting changes from GPT\n")

	prompt := p.Txt
	inFns := p.In
	outFns := p.Out
	var outFls []core.FileLang
	for _, fn := range outFns {
		lang, known, err := util.Ext2Lang(fn)
		Ck(err)
		if !known {
			Pf("Unknown language for file %s, defaulting to text\n", fn)
			lang = "text"
		}
		outFls = append(outFls, core.FileLang{File: fn, Language: lang})
	}

	sysmsg := "You are an expert Go programmer. Please make the requested changes to the given code."
	msgs := []core.ChatMsg{
		core.ChatMsg{Role: "USER", Txt: prompt},
	}

	resp, err := g.SendWithFiles(sysmsg, msgs, inFns, outFls)
	Ck(err)

	// ExtractFiles(outFls, promptFrag, dryrun, extractToStdout)
	err = core.ExtractFiles(outFls, resp, false, false)
	Ck(err)

	return
}

func getPrompt(promptfn string) (p *Prompt, err error) {
	defer Return(&err)
	var rc int

	// read or create the prompt file
	p, err = NewPrompt(promptfn)
	Ck(err)

	// create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	Ck(err)
	defer watcher.Close()
	// watch the prompt file
	err = watcher.Add(promptfn)
	Ck(err)

	// if AIDDA_EDITOR is set, open the editor where the users can
	// type a natural language instruction
	editor := envi.String("AIDDA_EDITOR", "")
	if editor != "" {
		Pf("Opening editor %s\n", editor)
		rc, err = RunInteractive(Spf("%s %s", editor, promptfn))
		Ck(err)
		Assert(rc == 0, "editor failed")
	}
	// wait for the file to be saved
	Pf("Waiting for file %s to be saved\n", promptfn)
	err = waitForFile(watcher, promptfn)
	Ck(err)

	// re-read the prompt file
	p, err = NewPrompt(promptfn)
	Ck(err)

	return p, err
}

func commit(g *core.Grokker) (err error) {
	defer Return(&err)
	var rc int
	// check git status for uncommitted changes
	stdout, stderr, rc, err := Run("git status --porcelain", nil)
	Ck(err)
	if len(stdout) > 0 {
		Pl(string(stdout))
		Pl(string(stderr))
		res, err := ask("There are uncommitted changes. Commit?", "y", "n")
		Ck(err)
		if res == "y" {
			// git add
			rc, err = RunInteractive("git add -A")
			Assert(rc == 0, "git add failed")
			Ck(err)
			// generate a commit message
			summary, err := g.GitCommitMessage("--staged")
			Ck(err)
			// git commit
			stdout, stderr, rc, err := Run("git commit -F-", []byte(summary))
			Assert(rc == 0, "git commit failed")
			Ck(err)
			Pl(string(stdout))
			Pl(string(stderr))
		}
	}
	return err
}

// getFiles returns a list of files to be processed
func getFiles() (files []string, err error) {
	defer Return(&err)
	// send that along with all files to GPT API
	// get ignore list
	ignore := []string{}
	ignorefn := ".aidda/ignore"
	if _, err := os.Stat(ignorefn); err == nil {
		// open the ignore file
		fh, err := os.Open(ignorefn)
		Ck(err)
		defer fh.Close()
		// read the ignore file
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			// split the ignore file into a list of patterns, ignore blank
			// lines and lines starting with #
			line := scanner.Text()
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				continue
			}
			ignore = append(ignore, line)
		}
	}

	// get list of files recursively
	files = []string{}
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		// ignore .git and .aidda directories
		if strings.Contains(path, ".git") || strings.Contains(path, ".aidda") {
			return nil
		}
		// check if the file is in the ignore list
		for _, pat := range ignore {
			re, err := regexp.Compile(pat)
			Ck(err)
			m := re.MatchString(path)
			if m {
				return nil
			}
		}
		// skip non-files
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		// add the file to the list
		files = append(files, path)
		return nil
	})
	Ck(err)
	return files, nil
}

// waitForFile waits for a file to be saved
func waitForFile(watcher *fsnotify.Watcher, fn string) (err error) {
	defer Return(&err)
	// wait for the file to be saved
	for {
		select {
		case event, ok := <-watcher.Events:
			Assert(ok, "watcher.Events closed")
			Pf("event: %v\n", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				Pf("modified file: %s\n", event.Name)
				// check if absolute path of the file is the same as the
				// file we are waiting for
				if filepath.Clean(event.Name) == filepath.Clean(fn) {
					Pf("file %s written to\n", fn)
					// wait for writes to finish
					time.Sleep(1 * time.Second)
					return
				}
			}
		case err, ok := <-watcher.Errors:
			Assert(ok, "watcher.Errors closed")
			return err
		}
	}
}
