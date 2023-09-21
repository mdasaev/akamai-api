# akamai-api 
## Script imports hostnames that have been activated on production properties and puts them on new version of Security Configuration

### NOTE: This works for WAP only without ASM

Script can be triggered manually but can be changed to other options, i.e by a push to main branch.
Comments in code provides the detailed workflow

To enable GH Actions:
 - Create an environment called "globaldots"
 - Add env variables from your .edgerc or IAM module :
    **AKAMAI_EDGEGRID_ACCESS_TOKEN, 
    AKAMAI_EDGEGRID_CLIENT_TOKEN,
    AKAMAI_EDGEGRID_CLIENT_SECRET,
    AKAMAI_EDGEGRID_HOST**
- Trigger manually to execute

To see what hostnames have been added check job log for output: "Hostnames to be added:"

### by DEFAULT new version will be activated on STAGING AKAMAI network

