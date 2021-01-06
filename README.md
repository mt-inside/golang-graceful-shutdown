Test cases
* Task fails to start: change one task's port to match the other
* Task errors at runtime: HTTP GET :8081/task1/quit
* HTTP server shutdown errors: change `if false` to `if true` in the `case <-stopCh:` branch of a task
  * To see it all work together, do this and then hit `/quit` on the other task
* Emergency shutdown: hit ^C twice in quick succession
