from steps.utils import send_get_request, send_delete_request, send_post_request


def setup_context_with_global_params_test(context):
    context.sites_to_be_cleaned_up = []


def create_site_entry(sites_names):
    endpoint = "dcim/sites/"
    for site in sites_names:
        payload = {
            "name": site,
            "slug": site.lower().replace(" ", "-"),
            "facility": "Omega",
            "description": "",
            "physical_address": "123 Fake St Lincoln NE 68588",
            "shipping_address": "123 Fake St Lincoln NE 68588",
            "comments": "Lorem ipsum etcetera",
        }
        send_post_request(payload, endpoint)


def remove_sites_entry(sites_names):
    endpoint = "dcim/sites/"
    for site in sites_names:
        site_id = (
            send_get_request(endpoint, {"name__ic": site})
            .json()
            .get("results")[0]
            .get("id")
        )
        send_delete_request(endpoint, site_id)


def before_all(context):
    setup_context_with_global_params_test(context)


def before_tag(context, tag):
    if tag == "update.object":
        create_site_entry(["Site-Test-2"])
    if tag == "object.state":
        create_site_entry(["Site Z", "Site X"])


def after_tag(context, tag):
    switcher = {
        "create.object": ["Site-Test"],
        "update.object": ["Site-Test-2"],
        "object.state": ["Site Z", "Site X"],
    }
    remove_sites_entry(switcher.get(tag, []))
