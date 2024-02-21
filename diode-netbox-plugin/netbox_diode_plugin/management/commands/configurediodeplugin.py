from django.contrib.auth.models import User
from django.core.management.base import BaseCommand
from users.models import NetBoxGroup, Token


def _create_user_with_token(username: str, api_key: str, group: NetBoxGroup, is_superuser: bool = False) -> None:
    """Create a user with the given username and API key if it does not exist."""
    try:
        user = User.objects.get(username=username)
    except User.DoesNotExist:
        if is_superuser:
            user = User.objects.create_superuser(username=username, is_active=True)
        else:
            user = User.objects.create(username=username, is_active=True)

    user.groups.add(*[group.id])

    if not Token.objects.filter(user=user).exists():
        Token.objects.create(user=user, key=api_key)


class Command(BaseCommand):
    """Configure NetBox Diode plugin."""

    help = "Configure NetBox Diode plugin"

    diode_to_netbox_username = "DIODE_TO_NETBOX"
    netbox_to_diode_username = "NETBOX_TO_DIODE"
    ingestion_username = "INGESTION"

    def add_arguments(self, parser):
        """Add command arguments."""
        parser.add_argument(
            "--diode-to-netbox-api-key",
            dest="diode_to_netbox_api_key",
            required=True,
            help="Diode to NetBox Diode plugin API key"
        )
        parser.add_argument(
            "--netbox-to-diode-api-key",
            dest="netbox_to_diode_api_key",
            required=True,
            help="NetBox Diode plugin to Diode API key"
        )
        parser.add_argument(
            "--ingestion-api-key",
            dest="ingestion_api_key",
            required=True,
            help="Ingestion API key"
        )

    def handle(self, *args, **options):
        """Handle command execution."""
        self.stdout.write("Configuring NetBox Diode plugin...")

        group, _ = NetBoxGroup.objects.get_or_create(name='diode')

        _create_user_with_token(self.diode_to_netbox_username, options['diode_to_netbox_api_key'], group)
        _create_user_with_token(self.netbox_to_diode_username, options['netbox_to_diode_api_key'], group, True)
        _create_user_with_token(self.ingestion_username, options['ingestion_api_key'], group)

        self.stdout.write("Finished.")
