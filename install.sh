#!/bin/bash

NAME="gocrawsan"
BIN_NAME="goc"
VERSION="0.1.0"
BIN_PREFIX="/usr/local"
COMP_PREFIX="/usr/local/share/zsh/site-functions"
GITHUB_USER="tzmfreedom"
TMP_DIR="/tmp"

set -ue

# check OS and architecture
UNAME=$(uname -s)
if [ "$UNAME" != "Linux" -a "$UNAME" != "Darwin" ] ; then
    echo "Sorry, OS not supported: ${UNAME}. Download binary from https://github.com/${USERNAME}/${NAME}/releases"
    exit 1
fi


if [ "${UNAME}" = "Darwin" ] ; then
  OS="darwin"

  OSX_ARCH=$(uname -m)
  if [ "${OSX_ARCH}" = "x86_64" ] ; then
    ARCH="amd64"
  else
    echo "Sorry, architecture not supported: ${OSX_ARCH}. Download binary from https://github.com/${USERNAME}/${NAME}/releases"
    exit 1
  fi
elif [ "${UNAME}" = "Linux" ] ; then
  OS="linux"

  LINUX_ARCH=$(uname -m)
  if [ "${LINUX_ARCH}" = "i686" ] ; then
    ARCH="386"
  elif [ "${LINUX_ARCH}" = "x86_64" ] ; then
    ARCH="amd64"
  else
    echo "Sorry, architecture not supported: ${LINUX_ARCH}. Download binary from https://github.com/${USERNAME}/${NAME}/releases"
    exit 1
  fi
fi


ARCHIVE_FILE=${NAME}-${VERSION}-${OS}-${ARCH}.tar.gz
BINARY="https://github.com/${GITHUB_USER}/${NAME}/releases/download/v${VERSION}/${ARCHIVE_FILE}"

cd $TMP_DIR
curl -sL -O ${BINARY}

tar xzf ${ARCHIVE_FILE}
mv ${OS}-${ARCH}/${BIN_NAME} ${BIN_PREFIX}/bin/${BIN_NAME}
chmod +x ${BIN_PREFIX}/bin/${BIN_NAME}
if [ -d ${COMP_PREFIX} ]; then
  echo "install zsh completion?(y/N): "
  read ln;
  if [ "${ln}" == "y" ]; then
    mv ${OS}-${ARCH}/_${BIN_NAME} ${COMP_PREFIX}/_${BIN_NAME}
  fi
fi

rm -rf ${OS}-${ARCH}
rm -rf ${ARCHIVE_FILE}
