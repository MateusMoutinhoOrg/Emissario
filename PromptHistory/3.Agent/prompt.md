
### objective:
create a configurable public agent, to talk with public users.



### features-font:
the root must be able to configure the agent-repo, 

### agent-repo-description:
its a github repo ,that its added with  a similar mechanic of the source repo, where the user can configure the agent instructions.

### agent-repo-structure:
- /instructions/ (dir with all the content the agent can ascess)
- /instructions/README.md (entry point of the agent system role)

### agent-functions-list-and-read-instructions:
the agent must be able to list, and retrive any file inside the instructions dir, 

 
### agent-functions-call-external-urls:
the agent must be able to call the urls configured in features-urls-acess with the tools get, and post with body.
 


### features-key-config:
the root user must be able to configure a api token key.

### features.max-daly-limit: 
the root must be able to configure the max cost per day.

### features-max-ip-daly-limit:
the root must be able to configure the max cost ip per day.

### features-model:
the root must be able to configure the model.

### features-urls-acess:
the user can configure urls that the agent can consult, and also configure predefined headers to pass to these urls.



