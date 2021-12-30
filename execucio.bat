@echo off

set abelles=Maya Reina Pana Willy Jota Zeta Alpha Omega Jane Monsa

FOR %%A IN (%abelles%) DO (
  go run send.go %%A
)


