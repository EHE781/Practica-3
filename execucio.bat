@echo off

::Lista de nombres de las abejas (determina cuantas abejas se inician)

set abelles=Maya Reina Pana Willy Jota Zeta Alpha Omega Jane

::set muchas=Carina Booker Jocelynn Ibarra Molly Lester Estrella Farmer Alejandra Mcpherson Gael Mclean Cristian Hartman Lyric Mcbride Antoine Burgess Zack Vasquez Priscilla Spears Naima Mathis Mary Steele Katrina Hartman Jayson Kramer Rachael Gentry Aisha Mcmahon Braelyn Munoz Noelle Tran Keyla Keller Anna Durham Deegan Clayton Morgan Lane Belinda Ewing Dangelo Rojas Danielle Mathews Dixie Newton Roger Castro Damarion Deleon Krish Hooper Jordyn Valencia Delaney Carpenter Hailey Durham

::Ejecución con una terminal universal (todos los prints en la misma, pueden haber errores visuales):

FOR %%A IN (%abelles%) DO (
  START/B go run abella/abella.go %%A
)

::Ejecución con una terminal por cada abeja:

::FOR %%A IN (%abelles%) DO (
  ::START/B go run abella/abella.go %%A
::)

