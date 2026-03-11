INSTALL_DIR := $(HOME)/.local/bin

.PHONY: install

install:
	go build -o $(INSTALL_DIR)/gh-dash .
