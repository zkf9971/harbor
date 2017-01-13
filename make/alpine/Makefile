# Makefile for a harbor project
#	
# Targets:
#
# build: 	build harbor photon images
# clean:	clean ui and jobservice harbor images

# common
SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKEDEVPATH=$(MAKEPATH)/dev
SRCPATH=./src
TOOLSPATH=$(BUILDPATH)/tools
CHECKENVCMD=checkenv.sh
DEVFLAG=true

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERIMASES=$(DOCKERCMD) images

# binary 
UISOURCECODE=$(SRCPATH)/ui
UIBINARYPATH=$(MAKEDEVPATH)/ui
UIBINARYNAME=harbor_ui

# alpine dockerfile
DOCKERFILEPATH=$(MAKEPATH)/alpine
DOCKERFILEPATH_UI=$(DOCKERFILEPATH)/ui
DOCKERFILENAME_UI=Dockerfile
DOCKERIMAGENAME_UI=vmware/harbor-ui

# version prepare
VERSIONFILEPATH=$(SRCPATH)/views/sections
VERSIONFILENAME=header-content.htm
GITCMD=$(shell which git)
GITTAG=$(GITCMD) describe --tags
ifeq ($(DEVFLAG), true)        
	VERSIONTAG=dev
else        
	VERSIONTAG=$(shell $(GITTAG))
endif

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

build: 
	@echo "building ui container for alpine..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_UI)/$(DOCKERFILENAME_UI) -t $(DOCKERIMAGENAME_UI):$(VERSIONTAG) .
	@echo "Done."
	
cleanimage:
	@echo "cleaning image for alpine..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	
.PHONY: clean
clean: cleanimage

