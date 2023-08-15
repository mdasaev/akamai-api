# akamai-api 
## Script imports hostnames that have been activated on production properties and puts them on Security Configuration

### NOTE: This works for WAP only without ASM

Script can be triggered manually but can be changed to other options, i.e by a push to main branch.
Comments in code provides the detailed workflow

To enable GH Actions:
 1 Create an environment called "globaldots"
 2 Add env variables from your .edgerc or IAM module :
  AKAMAI_EDGEGRID_ACCESS_TOKE: 
  AKAMAI_EDGEGRID_CLIENT_TOKEN 
  AKAMAI_EDGEGRID_CLIENT_SECRET
  AKAMAI_EDGEGRID_HOST
3 Trigger manually to execute
