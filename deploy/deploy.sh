BINARY=cloud

# Compile binary
cd ../code
make
mv $BINARY ../deploy
cd ../deploy

# Execute notebook
ansible-playbook playbook.yaml -i inventory.yaml --timeout 50
