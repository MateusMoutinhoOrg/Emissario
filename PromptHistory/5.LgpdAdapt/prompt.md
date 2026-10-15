
## Objective: 

Implement a security mechanism to ensure legal requirements its safe.

## Important:
contact ,ips, and summaries must be done on backoffice 


## Concat config 
if the root user allows contact, it needs to swhow a popup saying that the server will store user data, and these violates (puts laws that violates on the popup)

if the root user alloweed contacts,he needs to be able to set the max time that data will be stored in the database 
(the server needs to delete that information after that data)
(if that data is hgiher than some laws, it needs to show a popu saying it)


## Ip config 
if the user allows ip store , it needs to show a popup say that these will violates (puts the law on popup)


if the root user allows ip store, the user needs to be able to to set the max time that data will be stored 
(if that data is hgiher than some laws, it needs to show a popu saying it)

## Summaries 
if the user allows sumaries , it needs to shows a popup warning sayning that IA can make mistakes and its possible to store user data accidentally .


## default behavior 
- by default the contacts routes dont work 
- by default the ips is stored using a hash function (similar to password but way more faster)
- by default the sumarries route dont work