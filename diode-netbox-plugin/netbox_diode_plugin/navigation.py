from extras.plugins import PluginMenu, PluginMenuItem

# default arguments for the PluginMenuItem class
display_state = {
    "link": "plugins:netbox_diode_plugin:display_state",
    "link_text": "Display state",
    "staff_only": True                  # 3.6+ feature
}


menu = PluginMenu(
    label="NetBox Labs",
    groups=(
        (
            "Diode",
            (PluginMenuItem(**display_state),),
        ),
    ),
    icon_class="mdi mdi-arrow-collapse-right",
)
