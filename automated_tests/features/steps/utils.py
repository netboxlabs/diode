import requests

from features.steps.config import TestConfig

configs = TestConfig.configs()
api_root_path = str(configs["api_root_path"])
token = str(configs["user_token"])

headers = {
    "Content-Type": "application/json",
    "Authorization": f"Token {token}",
}


def send_request(payload):
    """Send a request to the API with the given payload and headers. Return the response."""
    endpoint = "plugins/diode/apply-change-set/"

    try:
        response = requests.post(
            f"{api_root_path}/{endpoint}", json=(payload), headers=headers
        )
    except Exception as e:
        print("Error:", str(e))
        return ValueError(e), None
    return response


#
#
# @given("that a valid {role} token exists")
# def test_authenticate(context, role):
#     context.token, id_token = get_valid_token(role)
#     assert_that(id_token, not_none(), f"Unable to get valid token. {context.token}")
#     context.role = role
#
#
# def get_valid_token(role):
#     if role == "admin":
#         token, id_token = authenticate_and_get_token_using_username_and_password(
#             configs.get("nbl_console_admin_username"),
#             configs.get("nbl_console_admin_password"),
#             configs.get("cognito_client_id"),
#         )
#     elif role == "unpriv":
#         token, id_token = authenticate_and_get_token_using_username_and_password(
#             configs.get("nbl_console_username"),
#             configs.get("nbl_console_password"),
#             configs.get("cognito_client_id"),
#         )
#     else:
#         raise ValueError("Unexpected token type. Options are 'admin' and 'unpriv'")
#     verify_token_route = f"{platform_api_root_path}auth/verifytoken/"
#     status_code, response = return_api_get_response(token, verify_token_route)
#     assert_that(status_code, equal_to(200), f"Failed to validate token: {response}")
#     return token, id_token
