* Properly implement and test compression plugins (gzip, bzip2, lzw, flate)
* Implement remaining unittests for compressions
    * p-lzma
    * p-gzip
    * p-lzw
    * p-bzip2
    * p-zlib
* Update CLI help infos for all plugins that add CLI arguments
* p-lzw: add LSB/MSB values to help
* Add CLI help infos, e. g. for plugins that use any RFC implementations
* Make sure process/unprocess functions are implemented for the applicable plugins
    * compression plugins
* Check p-flate code
* GUI
    * Implement Open/Save to file