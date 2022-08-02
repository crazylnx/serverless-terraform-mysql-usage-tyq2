locals {
  databases = [
    {
      "name" : "test_db",
      "character_set" : "utf8",
      "description" : "test database"
    },
  ]
}

provider "alicloud" {
  region = var.region != "" ? var.region : null
}


variable "region" {
  description = "resource region"
  type        = string
  default     = "cn-huhehaote"
}

variable "instance_name" {
  description = "RDS instance name"
  type        = string
  default     = "test"
}

variable "account_name" {
  description = "RDS instance user account name"
  type        = string
  default     = "user"
}

variable "password" {
  description = "RDS instance account password"
  type        = string
  default     = "ab12cd34"
}

variable "allocate_public_connection" {
  description = "Whether to allocate public connection for a RDS instance."
  type        = bool
  default     = true
}

variable "security_ips" {
  description = "List of IP addresses allowed to access all databases of an instance"
  type        = list(any)
  default     = ["0.0.0.0/0",]
}


variable "privilege" {
  description = "The privilege of one account access database."
  type        = string
  default     = "ReadWrite"
}


#variable "vswitch_id" {
#  type        = string
#  description = "The vswitch id of the RDS instance. If set, the RDS instance will be created in VPC, or it will be created in classic network."
#  default     = ""
#}

variable "databases" {
  description = "The database list, each database is a map, the map contains the following attributes: name, character_set, description, like `[{\"name\":\"test\",\"character_set\":\"utf8\",\"description\":\"test database\"},]`. It conflicts with `database_name`."
  type        = list(map(string))
  default     = []
}
variable "instance_type" {
  description = "The type of instance, more in https://help.aliyun.com/document_detail/276975.html?spm=5176.rdsbuy.0.tip.detail.40a5752fwc2ZQu"
  type        = string
  default     = "mysql.n1.micro.1"
}


data "alicloud_zones" "example" {
  available_resource_creation = "Rds"
}

resource "alicloud_vpc" "example" {
  vpc_name   = var.instance_name
  cidr_block = "172.16.0.0/16"
}


resource "alicloud_vswitch" "example" {
  vpc_id       = alicloud_vpc.example.id
  cidr_block   = "172.16.0.0/24"
  zone_id      = data.alicloud_zones.example.zones.0.id
  vswitch_name = var.instance_name
}


resource "alicloud_security_group" "group" {
  name   = var.instance_name
  vpc_id = alicloud_vpc.example.id
}

module "rds" {
  #  source                     = "github.com/kubevela-contrib/terraform-alicloud-rds"
  source                     = "./rds"
  engine                     = "MySQL"
  engine_version             = "8.0"
  instance_type              = var.instance_type
  instance_storage           = "20"
  instance_name              = var.instance_name
  account_name               = var.account_name
  password                   = var.password
  allocate_public_connection = var.allocate_public_connection
  security_ips               = var.security_ips
  databases                  = length(var.databases) != 0 ? var.databases : local.databases
  region                     = var.region
  privilege                  = var.privilege
  vswitch_id                 = alicloud_vswitch.example.id
}


output "RESOURCE_IDENTIFIER" {
  description = "The identifier of the resource"
  value       = module.rds.db_instance_id
}

output "DB_ID" {
  value       = module.rds.db_instance_id
  description = "RDS Instance ID"
}

output "ZONE_ID" {
  value = data.alicloud_zones.example.zones.0.id
}

output "DB_NAME" {
  value       = module.rds.this_db_instance_name
  description = "RDS Instance Name"
}
output "DB_USER" {
  value       = module.rds.this_db_database_account
  description = "RDS Instance User"
}
output "DB_PORT" {
  value       = module.rds.this_db_instance_port
  description = "RDS Instance Port"
}
output "DB_HOST" {
  value       = module.rds.this_db_instance_connection_string
  description = "RDS Instance Host"
}
output "DB_PASSWORD" {
  value       = var.password
  description = "RDS Instance Password"
}

output "DATABASE_NAME" {
  value       = module.rds.this_db_database_name
  description = "RDS Database Name"
}


output "VPC_ID" {
  value       = alicloud_vpc.example.id
  description = "vpc id"
}


output "VSWITCH_ID" {
  value       = alicloud_vswitch.example.id
  description = "VSwitch Id"
}

output "SECURITY_GROUP_ID" {
  value       = alicloud_security_group.group.id
  description = "security_group"
}
