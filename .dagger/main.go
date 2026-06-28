package main

import (
	"context"
	"dagger/traggo-mcp/internal/dagger"
)

type TraggoMcp struct{}

func (t *TraggoMcp) Build(
	ctx context.Context,
	// +default "."
	source *dagger.Directory,
	registerPassword *dagger.Secret,
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
		WithRegistryAuth("registry.rannes.dev", "christian@rannes.dev", registerPassword).
		Publish(ctx, "registry.rannes.dev/traggo-mcp:latest")

}
