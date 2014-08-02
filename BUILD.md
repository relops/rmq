Building rmq
------------

bintray
-------

To get goxc to publish to bintray, you need a .goxc.local.json with valid API credentials.

The file .goxc.local.json.example provides a template for this.

Cleaning Up
-----------

To delete previous versions on bintray:

	$ for v in ..; do http --auth USER:APIKEY DELETE https://api.bintray.com/packages/relops/rmq/rmq/versions/0.2.$v; done