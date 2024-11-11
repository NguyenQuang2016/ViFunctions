using MediatR;
using Microsoft.AspNetCore.Mvc;
using ViFunction.Orchestrator.Application.Commands;

namespace ViFunction.Orchestrator.Apis;

[Route("api/[controller]")]
[ApiController]
public class FunctionsController(IMediator mediator): ControllerBase
{
    [HttpPost("build")]
    public async Task<IActionResult> Build([FromForm] BuildCommand command)
    {
        var result = await mediator.Send(command);
        return result.IsSuccess ? Ok() : BadRequest(result);
    }
    [HttpPost("deploy")]
    public async Task<IActionResult> Deploy([FromBody] DeployCommand command)
    {
        var result = await mediator.Send(command);
        return result.IsSuccess ? Ok() : BadRequest(result);
    }
}