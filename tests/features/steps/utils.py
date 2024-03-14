import requests

from steps.config import TestConfig

configs = TestConfig.configs()
api_root_path = str(configs["api_root_path"])
token = str(configs["user_token"])

headers = {
    "Content-Type": "application/json",
    "Authorization": f"Token {token}",
}


def send_post_request(payload):
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


def send_get_request(endpoint, params=None):
    """Send a request to the API with the given endpoint and headers. Return the response."""
    try:
        if params:
            response = requests.get(
                f"{api_root_path}/{endpoint}", headers=headers, params=params
            )
        else:
            response = requests.get(f"{api_root_path}/{endpoint}", headers=headers)
    except Exception as e:
        print("Error:", str(e))
        return ValueError(e), None
    return response


def send_delete_request(endpoint, id):
    """Send a request to the API with the given endpoint and headers. Return the response."""
    try:
        response = requests.delete(f"{api_root_path}/{endpoint}/{id}/", headers=headers)
    except Exception as e:
        print("Error:", str(e))
        return ValueError(e), None
    return response
