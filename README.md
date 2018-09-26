CREATE TABLE key_value (
  name varchar(255) DEFAULT NULL,
  value bytea,
  revision bigint DEFAULT NULL,
  ttl bigint NOT NULL DEFAULT '0'
);

ALTER TABLE key_value ADD CONSTRAINT uix_key_value_name UNIQUE (name);
CREATE INDEX idx_key_value__ttl ON key_value (ttl);

go build

sample-apiserver --storage-backend=postgres \
--etcd-servers=postgres
--etcd-servers=k8s:k8s@tcp(localhost:3306)/k8s
--watch-cache=false