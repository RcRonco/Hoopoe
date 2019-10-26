# Rules
Rules defined in configuration file, their job is to Act in predefined action for specified DNS query.  
Every rule must be in the format: ```TYPE ACTION PATTERN OPTIONS```
Currently the are 4 types of rules supported.  
  
## Proxy Rules
* ```Pass``` - A rule is set for every query that the pattern matching to, will passed without any other rule type.
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None
         
* ```Allow``` - A Whitelist rule, any request that not match any ```Allow``` rule will be **DROPPED**.    
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None
         
* ```Deny``` - A Blacklist rule, any request that match one of the ```Deny``` rule will be **DROPPED**.   
    When ```Allow``` rule is also defined the Deny rule is used to block specific query inside the Whitelist query space.    
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None
         
* ```Rewrite``` - This rule used to edit the query before it arriving the Remote DNS Server.    
  **Parameters**:   
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: 
        * Replacement: ```string``` - string to replace pattern with.

## String Matching Actions
Currently the type of actions are ```String Matching```:
* **PREFIX**: Matching the prefix of string with ```Pattern```.
* **SUFFIX**: Matching the suffix of string with ```Pattern```.
* **SUBSTRING**: Will match if string contains ```Pattern```.
* **REGEXP**: Will match if string matches regexp ```Pattern```.