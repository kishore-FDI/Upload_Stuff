import requests
import jwt
import os
BASE_URL = "http://localhost:8080/api/v1"
JWT_SECRET = os.environ.get("JWT_Secret", "supersecretkey")
JWT_ALGORITHM = "HS256"

# 1. Sign Up
signup_data = {
    "username": "testbusiness",
    "password": "TestPassword123!",
    "email": "test@example.com"
}

resp = requests.post(f"{BASE_URL}/auth/signup", json=signup_data)
print("SignUp:", resp.status_code, resp.text)

# 2. Sign In
signin_data = {
    "username": "testbusiness",
    "password": "TestPassword123!"
}

resp = requests.post(f"{BASE_URL}/auth/signin", json=signin_data)
print("SignIn:", resp.status_code, resp.text)

if resp.status_code != 200:
    exit()

tokens = resp.json()
access_token = tokens["access_token"]
refresh_token = tokens["refresh_token"]

decoded = jwt.decode(access_token, JWT_SECRET, algorithms=[JWT_ALGORITHM])
business_username = decoded.get("username")
print("Logged in as business:", business_username)

headers = {"Authorization": f"Bearer {access_token}"}

# 3. Refresh Token
refresh_data = {"refresh_token": refresh_token}
resp = requests.post(f"{BASE_URL}/auth/refresh", json=refresh_data)
print("Refresh Token:", resp.status_code, resp.text)
if resp.status_code == 200:
    access_token = resp.json()["access_token"]
    headers = {"Authorization": f"Bearer {access_token}"}

# 4. Create API Key
resp = requests.post(f"{BASE_URL}/apikeys/create", headers=headers)
print("Create API Key:", resp.status_code, resp.text)
api_key = None
api_key_id = None
if resp.status_code == 200:
    api_key = resp.json()["api_key"]

# 5. List API Keys
resp = requests.get(f"{BASE_URL}/apikeys/list", headers=headers)
print("List API Keys:", resp.status_code, resp.text)

if resp.status_code == 200:
    keys = resp.json()
    # Find the key we just created
    for k in keys:
        if k["api_key"] == api_key:
            api_key_id = k["id"]
            break

# 6. Delete API Key
if api_key_id:
    resp = requests.delete(f"{BASE_URL}/apikeys/delete?id={api_key_id}", headers=headers)
    print("Delete API Key:", resp.status_code, resp.text)
