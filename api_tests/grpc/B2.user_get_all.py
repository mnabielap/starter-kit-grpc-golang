import sys
import os
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, load_config, GRPC_HOST, GRPC_PORT

from api.proto.v1 import user_pb2
from api.proto.v1 import user_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = user_pb2_grpc.UserServiceStub(channel)

token = load_config("accessToken")
metadata = [('authorization', f'Bearer {token}')]

payload = {
    "page": 1,
    "limit": 10,
    "sort": "created_at:desc"
}

run_grpc_test(
    stub=stub,
    method_name="ListUsers",
    request_proto_class=user_pb2.ListUsersRequest,
    payload=payload,
    metadata=metadata,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)