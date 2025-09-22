# using name
terraform import -var-file=../../terraform.tfvars catalyst_service_account.basic_service_account basic-service-account 
terraform import -var-file=../../terraform.tfvars catalyst_service_account.viewer_service_account readonly-service-account
terraform import -var-file=../../terraform.tfvars catalyst_service_account.cicd_service_account cicd-automation
