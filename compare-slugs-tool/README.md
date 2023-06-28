# Compare Slugs Tool
Purpose: To get a list of the Insights project slugs that have ES indices but those that don't exist in SFDC.

 - Get all the unique enriched ES indices. Except --for-merge and any index that is an alias
 - Get their project slugs from the index name
 - Check slug_mapping if the da_name exists in the table
 - List down the slugs for which slug_mapping entry does not exist.

## How to run the tool
 - clone v1 repo from <https://github.com/LF-Engineering/dev-analytics-api>
 - copy the fixtures files of the v1 repo from <https://github.com/LF-Engineering/dev-analytics-api/tree/prod/app/services/lf/bootstrap/fixtures> to `./compare-slugs-tool/fixtures`
 - run `$ go run main.go`
 - two files will be generated:
    * 1.`data_in_da_and_slug_mapping_not_in_sfdc.csv` will contain project data for projects in insights and in the slug mapping table but are NOT in SFDC.
    * 2.`data_not_in_slug_mapping_and_sfdc.csv` will contain project data for projects in insights fixtures but NOT in the slug mapping table and SFDC.

 - import the csv files into an Excel sheet for better readability

## Required Environment Variables

`ES_URL`= https://username:password@fqdn:port

`TOKEN` = auth0-token (Bearer + token)

`PROJECTS_SERVICE_BASE_URL` = https://api-gw.platform.linuxfoundation.org/project-service/v1/

`SH_DB` = username:password@tcp(host:port)/database_name?charset=utf8