import sys
import os
import grpc

sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import run_grpc_test, GRPC_HOST, GRPC_PORT

from api.proto.v1 import auth_pb2
from api.proto.v1 import auth_pb2_grpc

channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
stub = auth_pb2_grpc.AuthServiceStub(channel)

payload = {
    "email": "admin@example.com"
}

run_grpc_test(
    stub=stub,
    method_name="ForgotPassword",
    request_proto_class=auth_pb2.ForgotPasswordRequest,
    payload=payload,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)