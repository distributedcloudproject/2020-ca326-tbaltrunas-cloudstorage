BINARY=cloud

# Compile binary
cd ../code/cloud
make
mv $BINARY ../../deploy
cd ../../deploy

# Execute notebook
ansible-playbook playbook.yaml -i inventory.yaml --timeout 50
