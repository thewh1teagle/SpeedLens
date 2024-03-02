# Build

```console
name="thewh1teagle/speedlens"
docker build -t $name .
docker tag $name:latest $name:latest
docker push $name