import sys
import os
import time
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, save_config, GRPC_HOST, GRPC_PORT

# NESTED IMPORT
from api.proto.v1 import auth_pb2
from api.proto.v1 import auth_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = auth_pb2_grpc.AuthServiceStub(channel)

unique_id = int(time.time())
email = f"grpc_user_{unique_id}@example.com"

payload = {
    "name": "gRPC Test User",
    "email": email,
    "password": "password123"
}

response = run_grpc_test(
    stub=stub,
    method_name="Register",
    request_proto_class=auth_pb2.RegisterRequest,
    payload=payload,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)

if response.ok:
    data = response.json()
    save_config("accessToken", data['tokens']['access']['token'])
    save_config("refreshToken", data['tokens']['refresh']['token'])
    print(">>> Registration successful. Tokens saved.")