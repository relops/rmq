To delete previous versions on bintray:

	$ for v in ..; do http --auth USER:APIKEY DELETE https://api.bintray.com/packages/relops/rmq/rmq/versions/0.2.$v; done