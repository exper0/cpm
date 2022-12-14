# Container Process Monitor (CPM)
CPM is a lightweight process manager designed with containers in mind. Best practice is to have just one process per container but there are many cases when it's convenient or even required to have multiple processes within single container. In most cases something like shell script is being used in this case. However, shell script provides limited flexibility and is not too reliable when it comes to handling signals. Implementing some logic in shell script can also become quite messy very quickly. For instance, what do we do if one our two processes in container died? Should we kill container or just send notification somewhere? What if process simply exited with no errors? (e.g. provisioning script)
## Similar tools
Supervisord is what comes to mind, it's referenced in Docker documentation. CPM is more lightweight since it's written in Go and doesn't require having Python (or any other) interpreter in container. CPM also designed with slightly different use cases in mind

# Use cases
TODO
# Configuration
TODO
