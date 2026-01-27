//! Package test_module provides library functionality.

/// Returns a greeting message.
pub fn hello() -> &'static str {
    "Hello, World!"
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hello() {
        assert_eq!(hello(), "Hello, World!");
    }
}
