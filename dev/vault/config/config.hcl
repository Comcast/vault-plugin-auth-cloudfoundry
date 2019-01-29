log_level = "trace"
ui = true
cluster_name = "dev-server"
plugin_directory = "/var/vault/plugins"
disable_mlock = true

api_addr = "http://localhost:8200"

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_disable = 1
}

storage "inmem" {}
