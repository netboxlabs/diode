from behave import fixture, use_fixture

from steps.utils import send_get_request, send_delete_request


def setup_context_with_global_params_test(context):
    context.sites_to_be_cleaned_up = ["Site-Test"]


def delete_site_entry(site_name):
    endpoint = "dcim/sites/"
    site_id = (
        send_get_request(endpoint, {"name__ic": site_name})
        .json()
        .get("results")[0]
        .get("id")
    )
    response = send_delete_request(endpoint, site_id)
    return response


@fixture()
def site_cleanup(context):
    for site_name in context.sites_to_be_cleaned_up:
        delete_site_entry(site_name)


def before_all(context):
    setup_context_with_global_params_test(context)


def after_feature(context, feature):
    if "fixture.site.cleanup" in feature.tags:
        use_fixture(site_cleanup, context)
