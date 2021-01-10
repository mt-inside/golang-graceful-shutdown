## Test cases
* Task fails to start: change one task's port to match the other
* Task errors at runtime: HTTP GET :8081/task1/quit
* HTTP server shutdown errors: change `if false` to `if true` in the `case <-stopCh:` branch of a task
  * To see it all work together, do this and then hit `/quit` on the other task
* Emergency shutdown: hit ^C twice in quick succession

## Philosophy
This is classic programming-in-the-large.
Other approaches would be to
* Have the tasks isolated within the process somehow, and rather than report their errors, _Let It Crash_ (see Erlang) and restart. Go's errors-as-values makes this tricky but it can be done.
* Build the tasks into separate binaries and run in separate unix processes, then defer to something like Kubernetes as the supervisor, again Erlang-style.
