"""Utility functions for tests."""
import time
from typing import Callable, Any


def wait_for_condition(
    condition: Callable[[], bool],
    timeout: int = 30,
    interval: float = 1.0,
    error_message: str = "Condition not met within timeout"
) -> None:
    """Wait for a condition to become true.

    Args:
        condition: Callable that returns True when condition is met
        timeout: Maximum time to wait in seconds
        interval: Time between checks in seconds
        error_message: Error message if timeout is reached

    Raises:
        TimeoutError: If condition is not met within timeout
    """
    start_time = time.time()
    while time.time() - start_time < timeout:
        if condition():
            return
        time.sleep(interval)
    raise TimeoutError(error_message)


def retry_on_exception(
    func: Callable,
    max_retries: int = 3,
    delay: float = 1.0,
    exceptions: tuple = (Exception,)
) -> Any:
    """Retry a function call on exception.

    Args:
        func: Function to call
        max_retries: Maximum number of retry attempts
        delay: Delay between retries in seconds
        exceptions: Tuple of exceptions to catch

    Returns:
        Return value of the function

    Raises:
        The last exception if all retries fail
    """
    last_exception = None
    for attempt in range(max_retries):
        try:
            return func()
        except exceptions as e:
            last_exception = e
            if attempt < max_retries - 1:
                time.sleep(delay)
    raise last_exception