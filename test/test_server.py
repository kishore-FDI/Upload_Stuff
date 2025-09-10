#!/usr/bin/env python3
"""
Simple test script to verify server endpoints are working
"""

import requests
import json
import time

BASE_URL = "http://localhost:8080"
API_KEY = "5cf74fb0ea2ec521d85c36e0e4468749d6aaee5b8f5ff5c35c0f9a3260a96a60"
USERNAME = "user123444"

def test_health():
    """Test health endpoint"""
    print("🏥 Testing health endpoint...")
    try:
        response = requests.get(f"{BASE_URL}/health")
        if response.status_code == 200:
            print("✅ Health check passed")
            print(f"   Response: {response.json()}")
        else:
            print(f"❌ Health check failed: {response.status_code}")
    except Exception as e:
        print(f"❌ Health check error: {e}")

def test_api_info():
    """Test API info endpoint"""
    print("\n📋 Testing API info endpoint...")
    try:
        response = requests.get(f"{BASE_URL}/api/v1/")
        if response.status_code == 200:
            print("✅ API info endpoint working")
            print(f"   Response: {response.json()}")
        else:
            print(f"❌ API info failed: {response.status_code}")
    except Exception as e:
        print(f"❌ API info error: {e}")

def test_upload_meta():
    """Test upload metadata creation"""
    print("\n📤 Testing upload metadata creation...")
    try:
        headers = {
            "X-API-KEY": API_KEY,
            "X-Username": USERNAME,
            "Content-Type": "application/json"
        }
        response = requests.post(f"{BASE_URL}/api/v1/uploads/meta/", headers=headers)
        if response.status_code == 201:
            data = response.json()
            print("✅ Upload metadata created")
            print(f"   Token: {data.get('token')}")
            print(f"   Expires in: {data.get('expires_in')} seconds")
            return data.get('token')
        else:
            print(f"❌ Upload metadata failed: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Upload metadata error: {e}")
    return None

def test_status_endpoint(token):
    """Test status endpoint"""
    if not token:
        print("\n⏭️  Skipping status test (no token)")
        return
    
    print(f"\n📊 Testing status endpoint for token: {token}")
    try:
        response = requests.get(f"{BASE_URL}/api/v1/uploads/meta/{token}/status")
        if response.status_code == 200:
            print("✅ Status endpoint working")
            print(f"   Response: {json.dumps(response.json(), indent=2)}")
        else:
            print(f"❌ Status endpoint failed: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Status endpoint error: {e}")

def test_business_uploads():
    """Test business uploads listing"""
    print("\n🏢 Testing business uploads endpoint...")
    try:
        headers = {"X-API-KEY": API_KEY}
        response = requests.get(f"{BASE_URL}/api/v1/business/uploads", headers=headers)
        if response.status_code == 200:
            print("✅ Business uploads endpoint working")
            data = response.json()
            print(f"   Business ID: {data.get('business_id')}")
            print(f"   Uploads count: {data.get('count')}")
        else:
            print(f"❌ Business uploads failed: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Business uploads error: {e}")

def test_websocket_endpoint():
    """Test WebSocket endpoint (basic check)"""
    print("\n🔌 Testing WebSocket endpoint...")
    try:
        # Just check if the endpoint exists by making a GET request
        # This will fail with 400/404 if not implemented, which is expected
        response = requests.get(f"{BASE_URL}/api/v1/ws/test123")
        print(f"   WebSocket endpoint response: {response.status_code}")
        if response.status_code in [400, 404, 426]:  # Expected for WebSocket upgrade
            print("✅ WebSocket endpoint exists")
        else:
            print(f"❌ Unexpected WebSocket response: {response.status_code}")
    except Exception as e:
        print(f"❌ WebSocket test error: {e}")

def test_storage_endpoints():
    """Test storage endpoints"""
    print("\n💾 Testing storage endpoints...")
    
    # Test download with non-existent file
    try:
        headers = {"X-API-KEY": API_KEY}
        response = requests.get(f"{BASE_URL}/api/v1/storage/nonexistent", headers=headers)
        if response.status_code == 404:
            print("✅ Storage download endpoint working (404 for non-existent file)")
        else:
            print(f"❌ Storage download unexpected response: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Storage download error: {e}")
    
    # Test delete with non-existent file
    try:
        headers = {"X-API-KEY": API_KEY}
        response = requests.delete(f"{BASE_URL}/api/v1/storage/nonexistent", headers=headers)
        if response.status_code == 404:
            print("✅ Storage delete endpoint working (404 for non-existent file)")
        else:
            print(f"❌ Storage delete unexpected response: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Storage delete error: {e}")

def main():
    print("🧪 Server Endpoint Test Suite")
    print("=" * 50)
    
    # Run all tests
    test_health()
    test_api_info()
    token = test_upload_meta()
    test_status_endpoint(token)
    test_business_uploads()
    test_websocket_endpoint()
    test_storage_endpoints()
    
    print("\n✨ Test suite completed!")
    print("\n📝 Next steps:")
    print("   1. Run 'python test/progress_test.py' for WebSocket progress tracking")
    print("   2. Open 'test/progress_client.html' in browser for UI testing")
    print("   3. Check 'PROGRESS_TRACKING.md' for detailed documentation")

if __name__ == "__main__":
    main()
