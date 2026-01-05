import sys
import os
import time
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, load_config, save_config, GRPC_HOST, GRPC_PORT

# NESTED IMPORT for USER
from api.proto.v1 import user_pb2
from api.proto.v1 import user_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = user_pb2_grpc.UserServiceStub(channel)

token = load_config("accessToken")
if not token:
    print("Error: No access token found. Please login (A2) first.")
    sys.exit(1)

metadata = [('authorization', f'Bearer {token}')]

unique_id = int(time.time())
email = f"created_via_grpc_{unique_id}@example.com"

payload = {
    "name": "Created Via Python gRPC",
    "email": email,
    "password": "password123",
    "role": "user"
}

response = run_grpc_test(
    stub=stub,
    method_name="CreateUser",
    request_proto_class=user_pb2.CreateUserRequest,
    payload=payload,
    metadata=metadata,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)

if response.ok:
    data = response.json()
    save_config("target_user_id", data['id'])
    print(f">>> User created. ID {data['id']} saved.")