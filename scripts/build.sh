#!/usr/bin/env bash

XC_OS=$(go env GOOS)
XC_ARCH=$(go env GOARCH)
DEST_BIN=terraform-provider-softlayer

echo "Compiling for OS: $XC_OS and ARCH: $XC_ARCH"

gox -os="${XC_OS}" -arch="${XC_ARCH}" -output="${DEST_BIN}_{{.OS}}_{{.Arch}}" -ldflags="-s -w"

if [ $? != 0 ] ; then
    echo "Failed to compile, bailing."
    exit 1
fi

echo "Looking for Terraform install"

TERRAFORM=$(which terraform)

[ $TERRAFORM ] && TERRAFORM_LOC=$(dirname ${TERRAFORM})

if [ $TERRAFORM_LOC  ] ; then
    BASE_PATH=$TERRAFORM_LOC
else
    BASE_PATH=$GOPATH/bin
fi

echo ""
echo "Moving ${DEST_BIN}_${XC_OS}_${XC_ARCH} to $BASE_PATH/$DEST_BIN"
echo ""

mv ${DEST_BIN}_${XC_OS}_${XC_ARCH} $BASE_PATH/$DEST_BIN

echo "Resulting binary: "
echo ""
echo $(ls -la $BASE_PATH/$DEST_BIN)

# Conditional for Brew people on OSX
if [[ -L $TERRAFORM ]] && [[ $(uname) == "Darwin" ]]; then
  # If terraform is a symlink(probably installed by brew) then we want to link the binary there so its available to go at runtime.
  # Because readlink on OSX cannot fetch the absolute path, we used python.
  SYM_PATH=$(python -c "from __future__ import print_function ; import os ; print(os.path.realpath(\"${TERRAFORM}\"))")
  ln -sf $BASE_PATH/$DEST_BIN $(dirname $SYM_PATH)/$DEST_BIN
fi
