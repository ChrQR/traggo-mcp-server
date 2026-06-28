package main

import (
	"context"
	"dagger/traggo-mcp/internal/dagger"
)

type TraggoMcp struct{}

// Build compiles the server and publishes a container image to the given registry.
func (t *TraggoMcp) Build(
	ctx context.Context,
	// Source directory of the project.
	// +default "."
	source *dagger.Directory,
	// Container registry to publish to, e.g. "registry.example.com".
	registry string,
	// Username to authenticate against the registry.
	username string,
	// Password (or token) for the registry, passed as a Dagger secret.
	registryPassword *dagger.Secret,
	// Image repository and tag, appended to the registry host.
	// +default "traggo-mcp:latest"
	image string,
) (string, error) {

	buildCtr := dag.Container().From("golang:1.26.4-bookworm").
		WithEnvVariable("CGO_ENABLED", "0").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build-cache")).
		WithExec([]string{"go", "build", "-o", "./build/main", "./cmd/server"})

	finalCtr := dag.Container().From("cgr.dev/chainguard/wolfi-base").
		WithWorkdir("/app").
		WithFile("main", buildCtr.File("/app/build/main")).
		WithDirectory("./assets", source.Directory("/assets")).
		WithExposedPort(8080).
		WithEntrypoint([]string{"/app/main"})

	return finalCtr.
		WithRegistryAuth(registry, username, registryPassword).
		Publish(ctx, registry+"/"+image)

}
