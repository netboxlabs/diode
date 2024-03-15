from behave import fixture, use_fixture

from steps.utils import send_get_request, send_delete_request, send_post_request


def setup_context_with_global_params_test(context):
    context.sites_to_be_cleaned_up = []


def create_site_entry(context, sites_names):
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
        context.sites_to_be_cleaned_up.append(site)


@fixture()
def site_cleanup(context):
    endpoint = "dcim/sites/"
    for site_name_index in range(len(context.sites_to_be_cleaned_up)):

        site_id = (
            send_get_request(
                endpoint, {"name__ic": context.sites_to_be_cleaned_up[site_name_index]}
            )
            .json()
            .get("results")[0]
            .get("id")
        )
        send_delete_request(endpoint, site_id)
    context.sites_to_be_cleaned_up = []


def before_all(context):
    setup_context_with_global_params_test(context)


def before_feature(context, feature):
    if "fixture.create.site" in feature.tags:
        create_site_entry(context, ["Site-Test-2", "Site Z", "Site X"])


def after_feature(context, feature):
    if "fixture.site.cleanup" in feature.tags:
        if "fixture.create.site" in feature.tags:
            context.sites_to_be_cleaned_up.append("Site-Test")
        use_fixture(site_cleanup, context)
