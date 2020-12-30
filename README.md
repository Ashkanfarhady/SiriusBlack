# SiriusBlack
Sirius Black is a service which makes self destructing messages.
The original idea is from [Haphez](https://github.com/NimaGhaedsharafi/haphez).

[Sirius Black](https://en.wikipedia.org/wiki/Sirius_Black) from [Harry Potter series](https://en.wikipedia.org/wiki/Harry_Potter) was the Secret Keeper of the potters family, just like this service which keeps secrets.

Secrets can have TTL (in minutes) and a counter (how many times a secret can be read)

Example of setting the secret : 

```bash
curl 127.0.0.1:8080/create/ -X POST --data '{"body": "A top secret message to James Bond", "ttl":10, "counter": 2}' -H "content-type: application/json"
```

The above request will be respond with a uuid, for example : 
678fbc6b-1780-4d24-83e2-209ef434a396

Example of getitng the secret :

```bash
curl 127.0.0.1:8585/678fbc6b-1780-4d24-83e2-209ef434a396
```
