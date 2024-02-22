from dcim.models import Site
from django.contrib.contenttypes.models import ContentType
from extras.models import ObjectChange
from rest_framework import status
from utilities.testing import APITestCase


class ObjectStateListTestCase(APITestCase):

    @classmethod
    def setUpTestData(cls):
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

        site2_change = Site.objects.filter(id=2).update(name='Site 20')

        print(site2_change)

    def test_return_object_state_using_id(self):
        url = '/api/plugins/diode/object-state/?id=1&obj_type=dcim.site'
        response = self.client.get(url)
        self.assertEqual(response.json().get('object').get('name'), 'Site 1')
        self.assertEqual(response.status_code, status.HTTP_200_OK)

    def test_return_object_state_using_q(self):
        query_parameters = {
            "q": "Site 2",
            "obj_type": "dcim.site"
        }
        url = '/api/plugins/diode/object-state/?q=Site%202&obj_type=dcim.site'
        response = self.client.get(url)
        print(response.json())
        # self.assertEqual(response.json().get('object').get('name'), 'Site 2')
        self.assertEqual(response.status_code, status.HTTP_200_OK)
