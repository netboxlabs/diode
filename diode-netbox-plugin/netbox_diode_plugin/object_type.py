from django.contrib.contenttypes.models import ContentType


class SupportedObjectType:
    """Supported Object Types."""

    def __init__(self):
        """Initialize supported object types."""
        self.supported_object_types = {
            "dcim": [
                "device",
                "devicerole",
                "devicetype",
                "interface",
                "manufacturer",
                "platform",
                "site",
            ],
        }

    def get_supported_object_types(self, sender):
        """Get supported object types."""
        content_type = ContentType.objects.get_for_model(
            sender, for_concrete_model=False
        )
        app_label = content_type.app_label
        model_name = content_type.model

        return (
            f"{app_label}.{model_name}"
            if model_name in self.supported_object_types.get(app_label, [])
            else None
        )
