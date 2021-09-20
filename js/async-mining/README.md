This implements a readline loop to read terminal input. for each non-empty line it then starts a CPU-intensive process of trying to produce an sha256 checksum with 5 leading zeros in the hex representation by appending 20 random bytes, sort-of what bitcoin does.
Each job is voluntarily giving back control to the event loop after each 100000 hashes tried.
