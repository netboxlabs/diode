from dcim.models import Site
from django.contrib.auth import get_user_model
from django.test import TestCase
from users.models import Token

from netbox_diode_plugin.signals import handle_notify_diode

User = get_user_model()


class HandleNotifyDiodeTestCase(TestCase):
    """Test Handle Notify Diode class."""

    def setUp(self):
        """Set up."""
        self.user_netbox_to_diode = User.objects.create_user(username="NETBOX_TO_DIODE")
        Token.objects.create(user=self.user_netbox_to_diode)

        self.site = Site(
            id=10,
            name="Site 1",
            slug="site-1",
            facility="Alpha",
            description="First test site",
            physical_address="123 Fake St Lincoln NE 68588",
            shipping_address="123 Fake St Lincoln NE 68588",
            comments="Lorem ipsum etcetera",
        )
        self.site.save()

    def test_handle_notify_diode_success(self):
        """Test handle notify diode success."""
        instance = Site.objects.get(id=10)
        self.assertTrue(handle_notify_diode(instance, True, Site, None))

    def test_handle_notify_diode_failure(self):
        """Test handle notify diode failure."""
        Token.objects.filter(user=self.user_netbox_to_diode).delete()
        instance = Site.objects.get(id=10)
        self.assertFalse(handle_notify_diode(instance, True, Site, None))
