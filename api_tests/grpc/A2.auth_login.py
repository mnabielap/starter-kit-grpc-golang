import sys
import os
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, save_config, GRPC_HOST, GRPC_PORT

from api.proto.v1 import auth_pb2
from api.proto.v1 import auth_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = auth_pb2_grpc.AuthServiceStub(channel)

# Login as ADMIN to ensure B-Series tests pass
# Ensure this user exists in your DB!
payload = {
    "email": "admin@example.com", 
    "password": "password123" 
}

print(f">>> Attempting login for {payload['email']}...")

response = run_grpc_test(
    stub=stub,
    method_name="Login",
    request_proto_class=auth_pb2.LoginRequest,
    payload=payload,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)

if response.ok:
    data = response.json()
    save_config("accessToken", data['tokens']['access']['token'])
    save_config("refreshToken", data['tokens']['refresh']['token'])
    print(">>> Login successful. Tokens saved.")