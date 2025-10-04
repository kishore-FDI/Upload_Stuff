import requests
import json
print(requests.get("http://localhost:8080/health").text)