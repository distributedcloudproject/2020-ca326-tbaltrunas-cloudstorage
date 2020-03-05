STORAGEDIR=data
LOGSDIR=logs
KEYPATH=id_rsa
BINARYPATH=./cloud

$BINARYPATH -key $KEYPATH -name "Node 1" \
           -save-file save \
           -whitelist-file whitelist \
           -fancy-display -verbose \
           -log-level "DEBUG" -log-dir $LOGSDIR \
           -file-storage-capacity 100 -file-storage-dir $STORAGEDIR

$BINARYPATH -key $KEYPATH -name "Node 0" \
           -save-file save \
           -network "35.234.148.231:9000" \
           -fancy-display -verbose \
           -log-level "DEBUG" -log-dir $LOGSDIR \
           -file-storage-capacity 100 -file-storage-dir $STORAGEDIR

$BINARYPATH -key $KEYPATH -name "Node 2" \
           -save-file save \
           -network "35.234.148.231:9000" \
           -fancy-display -verbose \
           -log-level "DEBUG" -log-dir $LOGSDIR \
           -file-storage-capacity 100 -file-storage-dir $STORAGEDIR  \
           -file testfile


# clean up:
sudo rm -rf $LOGSDIR
sudo rm -rf $STORAGEDIR
sudo rm -rf $BINARYPATH
sudo rm -rf $KEYPATH
sudo rm -rf $KEYPATH.pub