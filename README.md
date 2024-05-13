# aidda

AI-Driven Development (AIDD) Assistant: uses
[grokker](https://github.com/stevegt/grokker) as a backend to support
a development cycle that looks like this:

- Write test cases, TDD style.  Add copious comments to the test
  cases, as these will be read by the LMM backend to help guide the
  development process.  These comments are where you express your
  intent about how the code should be written and behave beyond what
  can be explicitly tested -- shades of BDD. 
- Run aidda.  The tool will iteratively run the test cases, and use
  the test case output along with the comments to generate code that
  passes the test cases while converging on the intent expressed in
  the comments.  When the test cases pass, the tool will generate some
  recommendations for further development and then exit.

See the output of `aidda -h` for usage for now.

I use aidda every day.  It's a great way to at least get a first draft
of code written and explore solutions to a problem space.  I tend to
step into the code (still using github copilot) to work out details or
to get the algo unstuck when it writes itself into a corner. I've
found that the 128k-token version of GPT-4 is pretty capable otherwise
-- this won't work well with smaller token limits or earlier versions
of GPT.

One interesting thing about aidda is that it also does a pretty good
job of discovering and describing where the test cases themselves are
weak, ambiguous, or just plain wrong; this has always been a pitfall
with pure TDD or BDD.

At this point the tool is simply a shell script and is specific to
generating Go code on Linux.  I expect after the dust settles I'll
likely use aidda.sh to generate aidda.go for a compiled version of the
tool, and that will better enable the complexity needed to support
more capabilities. I'm open to pull requests, but am otherwise pretty
head-down actually using aidda as-is to work on some Go-specific
projects at the moment.

-- Steve
