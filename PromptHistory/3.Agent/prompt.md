
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


## features-resumes:
the server must be able to  store (the front-end-user decies) the resume of the conversation, these resumes, must have date , and the resume. The backoffice users, must be able to list,filter , search by date range ,and visualize each resume The root can delete the resumes.

## chatbot-workflow: 
- the users enters in the front-end route of chatbot (decided by the source repo )
- the front-end make the the call at /api/public-chat/menssages/ whith the user inputs
- the server adds the system instructions provided by the agent-repo, and the tool functions to retrive url,list agent-repo files,etc... 
- the servers uses the configured model to generate the next mensages
- the server respond the front with the new menssage
- the front adds the server response into the array of menssages and keeps the conversation.
- in the end the front-end decide if the conversation is stored in the resumes db. 
- if is to be stored the front make a call to  /api/public-chat/resumes/ passing the menssages,
- the server makes the resume using the configured model and store it.

## important:
- the front chat its configured by the root user of Emissario, so dont implement any part of the chat-bot, only provide the api routes. The only front-end that needs to be implemented its the features, since its backoffice area.


## routes-/api/public-chat/menssages/:
its recives the menssages array , that contains user inputs and server outputs
## routes-sample-/api/public-chat/menssages/: (post)
```json
[
    {
        "role": "user",
        "content": "Hello, how are you?"
    },
    {
        "role": "assistant",
        "content": "I am fine, thank you."
    }
]
```

## routes-sample-/api/public-chat/resume/ (post)
its recives the menssage array, and store the resume, its important to make a mandatory system instructions, to remove any particular user informations, such as names, dates, location, id, cpf, phone numbers etc.. 


