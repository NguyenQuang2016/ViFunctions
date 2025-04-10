using MediatR;
using ViFunction.Store.Application.Dtos;
using ViFunction.Store.Application.Entities;

namespace ViFunction.Store.Application.Requests;

public record GetFunctionByIdQuery(Guid Id) : IRequest<FunctionDto>;