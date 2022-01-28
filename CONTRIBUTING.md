# Contributing Guide

## Publish Module

1. Tidy up the module dependencies
   ```
   go mod tidy
   ```
2. Run the tests
     ```
     go test
     ```
3. Tag the change
   ```
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. Publish the new version
    ```
    GOPROXY=proxy.golang.org go list -m github.com/epiphyte/orchid@v0.1.0 
    ```