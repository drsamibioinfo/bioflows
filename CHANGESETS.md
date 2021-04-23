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
- Now, you can add config parameters to a tool or a pipeline as internal configuration parameters

# 0.0.3a Change Sets

- We have added the ability to mark a parameter in either Inputs and/or Config sections of a pipeline or a tool
 as attachable "attach: true" in the definition of a parameter in order to attach it to a container in case 
 the current tool will run in a container (To Be TESTED).
 
- We have added the ability to dynamically create new Steps parameters from within the JS script, 
You have to either use "self.the_name_of_the_variable = <Some Value> " or you can use it as

self.Add("variableName","<Variable Value>")
