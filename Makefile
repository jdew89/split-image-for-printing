ifneq (,)
.error This Makefile requires GNU Make.
endif

ifneq ("$(wildcard ./.env)","")
    include .env
	export
endif

sam_template_base=sam
sam_template=$(sam_template_base).yaml
sam_output=$(sam_template_base)-package.yaml

AWS_ACCOUNT      	?=641787608559
AWS_REGION       	?=us-east-1
AWS_PROFILE		 	?=dcrunch-prod

STAGE            	?=prod

FUNCTION_NAME_PREFIX ?=Copy-Models-From-Dev
FUNCTION_NAME		?=${FUNCTION_NAME_PREFIX}-${STAGE}

# Bucket where the SAM deloyment files are kept

S3_BUCKET 			?=dc.cloud.formation.${STAGE}

# can NOT contain periods, but can use hyphens
STACK_PREFIX     	?=Copy-Models-From-Dev
STACK_NAME       	?=${STACK_PREFIX}-${STAGE}

buildgo: 
	rm dist/main.zip & GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/main && zip -j dist/main.zip dist/main dist/index.html

build: buildgo
	sam build \
	    --template-file $(sam_template) \
		--region ${AWS_REGION}
# --use-container

# Building and deploying app
package: buildgo
	sam package \
	    --s3-bucket $(S3_BUCKET) \
		--s3-prefix ${FUNCTION_NAME_PREFIX} \
	    --template-file $(sam_template) \
	    --output-template-file $(sam_output) \
		--region ${AWS_REGION} \
		--profile ${AWS_PROFILE}

# Doesnt build and delpoy image
deploy: package
	sam deploy \
	    --template-file $(sam_output) \
		--s3-bucket $(S3_BUCKET) \
		--s3-prefix ${FUNCTION_NAME_PREFIX} \
	    --stack-name $(STACK_NAME) \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --region ${AWS_REGION} \
		--profile ${AWS_PROFILE} 
#		--confirm-changeset

azurepackage: 
	sam package \
	    --s3-bucket $(S3_BUCKET) \
		--s3-prefix ${FUNCTION_NAME_PREFIX} \
	    --template-file $(sam_template) \
	    --output-template-file $(sam_output) \
		--region ${AWS_REGION}

azuredeploy: 
	sam deploy \
	    --template-file $(sam_output) \
		--s3-bucket $(S3_BUCKET) \
		--s3-prefix ${FUNCTION_NAME_PREFIX} \
	    --stack-name $(STACK_NAME) \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --region ${AWS_REGION}

destroy: 
	sam delete --stack-name $(STACK_NAME)

forcedestroy:
	sam delete --stack-name $(STACK_NAME) --region ${AWS_REGION} --no-prompts

local:
	sam local start-api -t sam.yaml