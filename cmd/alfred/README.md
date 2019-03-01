# alfred

    go get github.com/codemodus/alfred

alfred is a CLI application for serving basic html/js/css sites. alfred follows
an opinionated directory structure and is well-suited to prototyping, but should
generally not be used for user-facing deployments.

    Available flags:

    --dir={somedir}     Set directory to serve.
    --port={:1234}      Set port to listen on.
    -acs                Set access logging.
    -s                  Set output silencing.

    Defaults:

    --dir=.
    --port=:4001

The directory structure used:
HTML files should be located in "{dir}/html".
Assets should be located in "{dir}/assets".
ICO files should be located in "{dir}/assets/ico".

For paths without extensions, an index.html file should be provided. For
example, if "localhost:4001/this" is requested, the following file should be
available "{dir}/html/this/index.html". 
