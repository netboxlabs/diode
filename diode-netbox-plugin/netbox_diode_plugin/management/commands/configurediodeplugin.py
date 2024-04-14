import os

from django.contrib.auth.models import User
from django.core.management.base import BaseCommand
from users.models import NetBoxGroup, Token


def _create_user_with_token(username: str, group: NetBoxGroup, is_superuser: bool = False) -> None:
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
        Token.objects.create(user=user, key=os.getenv(f"{username}_API_KEY"))


class Command(BaseCommand):
    """Configure NetBox Diode plugin."""

    help = "Configure NetBox Diode plugin"

    diode_to_netbox_username = "DIODE_TO_NETBOX"
    netbox_to_diode_username = "NETBOX_TO_DIODE"
    ingestion_username = "INGESTION"

    def handle(self, *args, **options):
        """Handle command execution."""
        self.stdout.write("Configuring NetBox Diode plugin...")

        group, _ = NetBoxGroup.objects.get_or_create(name="diode")

        _create_user_with_token(self.diode_to_netbox_username, group)
        _create_user_with_token(self.netbox_to_diode_username, group, True)
        _create_user_with_token(self.ingestion_username, group)

        self.stdout.write("Finished.")
