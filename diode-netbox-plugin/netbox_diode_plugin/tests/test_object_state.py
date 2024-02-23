from dcim.models import Site
from django.core.management import call_command
from rest_framework import status
from utilities.testing import APITestCase


class ObjectStateListTestCase(APITestCase):
    def setUp(self):
        # Create sites with a value for each cacheable field defined on SiteIndex
        sites = (
            Site(
                id=1,
                name='Site 1',
                slug='site-1',
                facility='Alpha',
                description='First test site',
                physical_address='123 Fake St Lincoln NE 68588',
                shipping_address='123 Fake St Lincoln NE 68588',
                comments='Lorem ipsum etcetera'
            ),
            Site(
                id=2,
                name='Site 2',
                slug='site-2',
                facility='Bravo',
                description='Second test site',
                physical_address='725 Cyrus Valleys Suite 761 Douglasfort NE 57761',
                shipping_address='725 Cyrus Valleys Suite 761 Douglasfort NE 57761',
                comments='Lorem ipsum etcetera'
            ),
            Site(
                id=3,
                name='Site 3',
                slug='site-3',
                facility='Charlie',
                description='Third test site',
                physical_address='2321 Dovie Dale East Cristobal AK 71959',
                shipping_address='2321 Dovie Dale East Cristobal AK 71959',
                comments='Lorem ipsum etcetera'
            ),
        )
        Site.objects.bulk_create(sites)

        # call_command is because the searching using q parameter uses CachedValue to get the object ID
        call_command('reindex')

        self.url = '/api/plugins/diode/object-state/'

    def test_return_object_state_using_id(self):
        query_parameters = {
            "id": 1,
            "obj_type": "dcim.site"
        }

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get('object').get('name'), 'Site 1')

    def test_return_object_state_using_q(self):
        query_parameters = {
            "q": "Site 2",
            "obj_type": "dcim.site"
        }

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get('object').get('name'), 'Site 2')

    def test_object_not_found_return_empty(self):
        query_parameters = {
            "q": "Site 10",
            "obj_type": "dcim.site"
        }

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json(), [])

    def test_missing_obj_type_return_400(self):
        query_parameters = {
            "q": "Site 10",
            "obj_type": ""
        }

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_missing_q_and_id_parameters_return_400(self):
        query_parameters = {
            "obj_type": "dcim.site"
        }

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
