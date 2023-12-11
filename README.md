# Warp-GPT
作为刚学go的一个练手项目，自用

- 将chatgpt前端进行逆向，实现绕过cloudflare
- 对官方api进行代理
- 实现前端接口转标准api(通过access_token实现标准api传入访问)

端口列表
```
/backend-api/* (前端逆向接口)
/api/* (前端逆向接口)
/public-api/* (前端逆向接口)
/v1/* (官方api代理)
/r/v1/chat/completions (前端接口转标准api,支持流式)
/r/v1/images/generations (前端接口转标准api,不支持流式,只支持gpt-4的账户)
/getsession (实现__Secure-next-auth.session-token刷新session，返回session,或输入username与password输出session)
/token (获取ArkoseToken)
```
```
/r/v1/chat/completions method:["GET", "POST", "OPTIONS"]
input:
{
  "model": "gpt-3.5-turbo-16k",
  "messages": [
    {
      "role": "user",
      "content": "what can you do"
    }
  ]
}
output:
{
    "id": "chatcmpl-m3mYrjKTZuoNARfQerON95UJlA9XSWBi",
    "object": "chat.completion",
    "created": 1701011706,
    "model": "gpt-3.5-turbo-16k",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "I can do a wide range of tasks and provide information on various topics. Here are some of the things I can do:\n\n1. Answer Questions: I can provide information on a wide range of topics, including science, history, technology, mathematics, and more.\n\n2. Generate Text: I can generate text for various purposes, such as writing essays, creating stories, composing emails, and more.\n\n3. Language Translation: I can translate text from one language to another.\n\n4. Math Assistance: I can help with mathematical calculations, equations, and explanations.\n\n5. Programming Help: I can assist with coding and programming-related questions and problems.\n\n6. Writing Assistance: I can help with grammar and writing suggestions, including editing and proofreading.\n\n7. General Knowledge: I can provide general knowledge and facts on a wide variety of subjects.\n\n8. Recommendations: I can offer recommendations for books, movies, music, travel destinations, and more.\n\n9. Conversation and Chat: I can engage in casual conversation and chat on a variety of topics.\n\n10. Learning and Education: I can assist with learning and provide explanations on academic subjects.\n\n11. Problem Solving: I can help you brainstorm ideas, solve problems, and make decisions.\n\n12. Trivia and Quizzes: I can create and participate in trivia quizzes and answer trivia questions.\n\nPlease keep in mind that I do not have access to real-time information beyond my last knowledge update in January 2022, so some information may be outdated, and I cannot provide current news or events. If you have a specific task or question in mind, feel free to ask, and I'll do my best to assist you!"
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 0,
        "completion_tokens": 0,
        "total_tokens": 0
    }
}
```
```
/r/v1/images/generations method:["GET", "POST", "OPTIONS"]
input:
{
  "model": "dall-e-3",
  "prompt": "A cute baby sea otter",
  "n": 1,
  "size": "1024x1024"
}
output:
{
  "created": 1701014049,
  "data": [
    {
      "revised_prompt": "A cute baby sea otter, looking fluffy and adorable, with big, curious eyes, floating on its back in a calm blue ocean. The otter is holding a small shell in its tiny paws, and its fur is wet, giving it a shiny appearance under the sun. The background features a serene seascape with gentle waves and a clear sky.",
      "url": "https://files.oaiusercontent.com/file-fqEmsHBijHGkBKo0CnYIAfCJ?se=2023-11-26T16%3A54%3A09Z&sp=r&sv=2021-08-06&sr=b&rscc=max-age%3D31536000%2C%20immutable&rscd=attachment%3B%20filename%3Daa87dac2-8142-419d-9fe1-afa90c0a376e.webp&sig=xjwmZhzC3fZSF7V6TJ5hTWkmxBOMiVQKs0v/wTJRvAM%3D"
    }
  ]
}
```
```
/getsession methods:["POST"]
input:
{
	"refreshCookie":""
}
or
{
	"username":"",
  "password":""
}
output:
{
    "user": {
        "id": "",
        "name": "",
        "email": "",
        "image": "",
        "picture": "",
        "idp": "auth0",
        "iat": 1701014297,
        "mfa": false,
        "groups": [],
        "intercom_hash": ""
    },
    "expires": "2024-02-24T15:58:17.821Z",
    "accessToken": "",
    "authProvider": "auth0",
    "models": [
        {
            "slug": "text-davinci-002-render-sha",
            "max_tokens": 8191,
            "title": "Default (GPT-3.5)",
            "description": "Our fastest model, great for most everyday tasks.",
            "tags": [
                "gpt3.5"
            ],
            "capabilities": {},
            "product_features": {}
        }
    ],
    "refreshCookie": ""
}
```
```
/token methods:["GET"]
output:
{
    "token": ""
}
```

## 使用方法
### 配置文件
- 在harPool目录中加入har文件，实现登录验证与gpt4对话验证
- 将.env.temp修改为.env，修改配置项后保存
``` python
proxy = "http://127.0.0.1:10809"
port = 5000
host = '127.0.0.1'
verify = false
auth_key = ""
arkose_must = false
OpenAI_HOST = "chat.openai.com"
proxy_pool_url=""
log_level = "debug"

redis_address = "127.0.0.1:6379"
redis_passwd = ""
redis_db = 0
```
其中proxy_pool_url使用的是[ipidea](https://share.ipidea.net/8hPKah)的代理池，注册送100M流量，无限ip，一个月，测试足够
使用代理池后需要填写redis信息，redis版本需要7以上
### 运行

`go build && ./WarpGPT`
