import time
from typing import Optional, Dict, Any

import requests
from netboxlabs.diode.sdk import DiodeClient
from steps.config import TestConfig

configs = TestConfig.configs()
api_root_path = str(configs["api_root_path"])
token = str(configs["user_token"])

headers = {
    "Content-Type": "application/json",
    "Authorization": f"Token {token}",
}


def send_post_request(payload, endpoint="plugins/diode/apply-change-set/"):
    """Send a request to the API with the given payload and headers. Return the response."""
    try:
        response = requests.post(
            f"{api_root_path}/{endpoint}", json=payload, headers=headers
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


def get_site_id(site_name):
    """Get the site ID by name."""
    endpoint = "dcim/sites/"
    site_id = (
        send_get_request(endpoint, {"name__ic": site_name})
        .json()
        .get("results")[0]
        .get("id")
    )
    return site_id


def get_object_by_name(name, endpoint):
    """Get the object by name."""
    response = send_get_request(endpoint, {"name__ic": name}).json().get("results")
    if response:
        return response[0]
    return None


def get_object_by_model(model, endpoint):
    """Get the object by model."""
    response = send_get_request(endpoint, {"model__ic": model}).json().get("results")
    if response:
        return response[0]
    return None


def get_object_state(params: dict, max_retries: int = 3) -> Optional[Dict[str, Any]]:
    """Get object using given endpoint and params."""
    endpoint = f"plugins/diode/object-state/"

    attempt = 0
    while attempt < max_retries:
        response = send_get_request(endpoint, params).json().get("object")
        if response:
            return response
        time.sleep(1)
        attempt += 1

    return None


def ingester(entities):
    """Ingest the site object using the Diode SDK"""
    api_key = str(configs["api_key"])
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        api_key=api_key,
    ) as client:
        entities = entities
        response = client.ingest(entities=entities)
        return response
