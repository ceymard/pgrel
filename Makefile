.PHONY: all

all:
	go build

.PHONY: addlicense
addlicense:
	addlicense -c "Christophe Eymard" .

