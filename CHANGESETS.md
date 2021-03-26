# 0.0.2a Change Sets 
- The ability to define general inputs to the pipeline definition.
- Added the ability to execute scripts in parent pipeline.
- General Bug Fixes and Improvements.

# 0.0.2b Change Sets
- The ability of adding arbitrary parameters to the command line of tool or pipeline
- The ability to define steps as loops within a pipeline only.
- Added the functionality to reference local files in URL directives instead of only remote http files.
- Added the functionality to reference local files in scripts directive as well.
- ListDir(DirPath : string , [Absolute: bool] )
  Absolute : is an optional boolean parameter indicates whether the file name returned is relative or absolute
  default : false
- Added the ability to make a single Tool as a loop. Previously,
 you could only allow a step in a PIPELINE to be a loop but now you 
 can allow a single Tool "type: tool" to be a loop and execute the tool like
 e.g: bf Tool run ......
