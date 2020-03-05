apt-get install postgresql-client postgresql
update-rc.d postgreql enable
systemctl start postgresql

# in /etc/postgresql/X/main/pg_hba.conf change postgres auth method from "peer" to "trust", and others to "md5"
systemctl restart postgresql

# In psql:
CREATE DATABASE cloud;
CREATE USER cloud_admin WITH ENCRYPTED PASSWORD 'yourpass';
GRANT ALL PRIVILEGES ON DATABASE cloud TO cloud_admin;

# Create table
psql -f cloud.sql cloud

# Access
psql -U cloud_admin -W cloud

# Check tables: \dt
# SELECT * FROM accounts;
# INSERT INTO accounts VALUES ('test', '$2y$08$PmEZA94zFbTS.vxyj2GMC.8m.2QZfIIokWCKS9D7XpPKqhQVsXmj6');
# UPDATE accounts SET password = 'PWHASHHERE';
# DELETE FROM accounts WHERE username='sampleuser';

# 'test', 'helloworld'
# 'bozh', 'birdclassifier'
# 'bien', 'failedproject'
# 'kyle', 'backgroundmodel'
