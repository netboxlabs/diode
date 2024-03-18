import configparser


class TestConfig:
    """A class to read the test configurations from the configs.ini file"""

    _configs = None

    def __init__(self):
        raise RuntimeError("Call instance() instead")

    @classmethod
    def configs(cls):
        if cls._configs is None:
            cls._configs = _read_configs()
        return cls._configs


def _read_configs():
    """Read the test configurations from the configs.ini file"""
    parser = configparser.ConfigParser()
    parser.read("./features/configs.ini")
    configs = parser["tests_config"]

    configs["api_root_path"] = configs.get("api_root_path", "http://0.0.0.0:8000/")
    configs["user_token"] = configs.get(
        "user_token",
    )

    return configs


configs = TestConfig.configs()
