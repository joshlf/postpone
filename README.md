postpone
========

The postpone package provides an io.ReadSeeker wrapper with extra functionality. Some examples:

Create a ReadSeeker for a file which only opens the file once the Read method is called.

Create a ReadSeeker for a file which preloads the file into RAM and closes the connection to speed file reads.