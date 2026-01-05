import sys
import os
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, save_config, load_config, GRPC_HOST, GRPC_PORT

from api.proto.v1 import auth_pb2
from api.proto.v1 import auth_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = auth_pb2_grpc.AuthServiceStub(channel)

refresh_token = load_config("refreshToken")

payload = {
    "refresh_token": refresh_token
}

response = run_grpc_test(
    stub=stub,
    method_name="RefreshToken",
    request_proto_class=auth_pb2.RefreshTokenRequest,
    payload=payload,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)

if response.ok:
    data = response.json()
    save_config("accessToken", data['access']['token'])
    save_config("refreshToken", data['refresh']['token'])
    print(">>> Tokens refreshed.")